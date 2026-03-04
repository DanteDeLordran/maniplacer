package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates Maniplacer to the latest version",
	Long: `Checks GitHub for the latest release of Maniplacer and updates the local binary if a newer version is available.

By default, the command will ask for confirmation before updating.
You can skip confirmation with the --force flag.

The update process:
1. Fetches the latest release version from GitHub.
2. Compares it with the currently installed version.
3. If a newer version exists, downloads the appropriate binary for your OS/ARCH.
4. Creates a backup of the existing binary.
5. Replaces the old binary with the new one via an update script (applied after the process exits).

Example:
  maniplacer update
  maniplacer update --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.LoggerFromContext(cmd.Context())

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			logger.Debug("could not parse force flag, using default", "error", err)
			force = false
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		version, err := getLatestVersion(ctx)
		if err != nil {
			return fmt.Errorf("could not get latest version: %w", err)
		}

		if version == utils.Version {
			fmt.Println("No new version available")
			return nil
		}

		logger.Info("new version available", "current", utils.Version, "latest", version)
		fmt.Println("New version available:", version)

		if !force {
			choice := utils.ConfirmMessage("Are you sure you want to update?")

			if !choice {
				fmt.Printf("Not updating, staying in version %s\n", utils.Version)
				return nil
			}
		}

		fmt.Printf("Updating from %s to %s...\n", utils.Version, version)

		if err := downloadAndReplace(ctx, version); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		logger.Info("update successful", "version", version)
		fmt.Println("Successfully updated to", version)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolP("force", "f", false, "Forces auto update")
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func getLatestVersion(ctx context.Context) (string, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/dantedelordran/maniplacer/releases/latest", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add User-Agent header (required by GitHub API)
	req.Header.Set("User-Agent", "maniplacer/"+utils.Version)

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to check releases: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", res.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(res.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode release info: %w", err)
	}

	version := release.TagName
	if len(version) > 0 && version[0] == 'v' {
		version = version[1:]
	}

	return version, nil
}

func downloadAndReplace(ctx context.Context, version string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	goos := runtime.GOOS
	arch := runtime.GOARCH
	binaryName := fmt.Sprintf("maniplacer-%s-%s", goos, arch)

	downloadURL, err := getDownloadURL(ctx, version, binaryName)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	utils.Logger().Info("downloading binary", "url", downloadURL)
	fmt.Printf("Downloading %s...\n", downloadURL)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute, // Longer timeout for downloads
	}

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", res.StatusCode)
	}

	tempFile, err := os.CreateTemp("", "maniplacer-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, res.Body)
	if err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tempFile.Close()

	if err := os.Chmod(tempFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make temp file executable: %w", err)
	}

	backupPath := execPath + ".backup"
	updatePath := execPath + ".update"
	scriptPath := execPath + ".update.sh"

	if err := copyFile(tempFile.Name(), backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	if err := copyFile(tempFile.Name(), updatePath); err != nil {
		return fmt.Errorf("failed to create update file: %w", err)
	}

	return replaceBinary(execPath, updatePath, backupPath, scriptPath)
}

func replaceBinary(execPath, updatePath, backupPath, scriptPath string) error {
	scriptContent := fmt.Sprintf(`#!/bin/bash
set -e
echo "Waiting for old process to exit..."

while lsof "%[1]s" &>/dev/null; do
    sleep 1
done

echo "Replacing old binary..."
mv "%[1]s" "%[2]s" 2>/dev/null || true
mv "%[3]s" "%[1]s"
chmod +x "%[1]s"

echo "Cleaning up..."
rm -f "%[2]s" "%[3]s" "%[4]s"

echo "Update complete."
`, execPath, backupPath, updatePath, scriptPath)

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		os.Remove(updatePath)
		return fmt.Errorf("failed to create update script: %w", err)
	}

	cmd := exec.Command("/bin/bash", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the script and exit immediately
	if err := cmd.Start(); err != nil {
		os.Remove(scriptPath)
		return fmt.Errorf("failed to start update script: %w", err)
	}

	fmt.Println("Update will complete after the program exits...")
	os.Exit(0)
	return nil
}

func getDownloadURL(ctx context.Context, version, binaryName string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	url := fmt.Sprintf("https://api.github.com/repos/dantedelordran/maniplacer/releases/tags/%s", version)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "maniplacer/"+utils.Version)

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get release info: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", res.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(res.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode release info: %w", err)
	}

	for _, asset := range release.Assets {
		if asset.Name == binaryName {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("binary %s not found in release assets", binaryName)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}
