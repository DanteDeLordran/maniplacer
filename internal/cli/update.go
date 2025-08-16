package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "A brief description of your command",
	Long:  `A`,
	Run: func(cmd *cobra.Command, args []string) {

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			fmt.Printf("Could not parse force flag, using default value %s\n", err)
			force = false
		}

		version, err := getLatestVersion()
		if err != nil {
			fmt.Println("Could not get latest version due to:", err)
			os.Exit(1)
		}

		if version == utils.Version {
			fmt.Println("No new version available")
			os.Exit(1)
		} else {
			fmt.Println("New version available:", version)
		}

		if !force {
			choice := utils.ConfirmMessage("Are you sure you want to update?")

			if !choice {
				fmt.Printf("Not updating, staying in version %s\n", utils.Version)
				os.Exit(1)
			}

		}
		fmt.Printf("Updating from %s to %s...\n", utils.Version, version)
		if err := downloadAndReplace(version); err != nil {
			fmt.Printf("Update failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully updated to", version)

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

func getLatestVersion() (string, error) {
	res, err := http.Get("https://api.github.com/repos/dantedelordran/maniplacer/releases/latest")
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

func downloadAndReplace(version string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	goos := runtime.GOOS
	arch := runtime.GOARCH
	binaryName := fmt.Sprintf("maniplacer-%s-%s", goos, arch)

	downloadURL, err := getDownloadURL(version, binaryName)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	fmt.Printf("Downloading %s...\n", downloadURL)
	res, err := http.Get(downloadURL)
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

# Optional: Uncomment to restart
# exec "%[1]s" "$@"
`, execPath, backupPath, updatePath, scriptPath)

	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		os.Remove(updatePath)
		return fmt.Errorf("failed to create update script: %w", err)
	}

	cmd := exec.Command("/bin/bash", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()

	fmt.Println("Update will complete after the program exits...")
	os.Exit(0)
	return nil
}

func getDownloadURL(version, binaryName string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/dantedelordran/maniplacer/releases/tags/%s", version)
	res, err := http.Get(url)
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
