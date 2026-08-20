package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

const (
	latestReleaseURL = "https://api.github.com/repos/dantedelordran/maniplacer/releases/latest"
	maxBinarySize    = 512 << 20
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Maniplacer to the latest version",
	Long: `Checks GitHub for a newer Maniplacer release, verifies the downloaded
binary against the SHA-256 digest published by GitHub, and atomically replaces
the current executable.

Self-update is supported on Linux and macOS. On Windows, rerun installer.sh so
the running executable can be replaced safely.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("self-update is not supported on Windows; rerun installer.sh to update")
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return fmt.Errorf("could not parse force flag: %w", err)
		}

		lookupCtx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		release, err := getLatestRelease(lookupCtx)
		cancel()
		if err != nil {
			return fmt.Errorf("could not get latest release: %w", err)
		}

		latest := normalizeVersion(release.TagName)
		newer, err := isNewerVersion(utils.Version, latest)
		if err != nil {
			return err
		}
		if !newer {
			fmt.Println("No new version available")
			return nil
		}

		fmt.Printf("New version available: %s\n", latest)
		if !force && !utils.ConfirmMessage("Are you sure you want to update?") {
			fmt.Printf("Not updating, staying on version %s\n", utils.Version)
			return nil
		}

		downloadCtx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
		defer cancel()
		if err := downloadAndReplace(downloadCtx, release); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		fmt.Println("Successfully updated to", latest)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolP("force", "f", false, "Skip update confirmation")
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		Digest             string `json:"digest"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func getLatestRelease(ctx context.Context) (GitHubRelease, error) {
	var release GitHubRelease
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return release, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "maniplacer/"+utils.Version)

	res, err := updateHTTPClient(30 * time.Second).Do(req)
	if err != nil {
		return release, fmt.Errorf("failed to check releases: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return release, fmt.Errorf("GitHub API returned status %d", res.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(&release); err != nil {
		return release, fmt.Errorf("failed to decode release info: %w", err)
	}
	if release.TagName == "" {
		return release, fmt.Errorf("release has no tag")
	}
	return release, nil
}

func updateHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			if req.URL.Scheme != "https" || !trustedDownloadHost(req.URL) {
				return fmt.Errorf("refusing redirect to untrusted URL %s", req.URL.Redacted())
			}
			return nil
		},
	}
}

func trustedDownloadHost(downloadURL *url.URL) bool {
	host := strings.ToLower(downloadURL.Hostname())
	return host == "github.com" || host == "api.github.com" ||
		strings.HasSuffix(host, ".githubusercontent.com")
}

func downloadAndReplace(ctx context.Context, release GitHubRelease) error {
	binaryName, err := releaseBinaryName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	var downloadURL, expectedDigest string
	for _, asset := range release.Assets {
		if asset.Name == binaryName {
			downloadURL = asset.BrowserDownloadURL
			expectedDigest = strings.TrimPrefix(strings.ToLower(asset.Digest), "sha256:")
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("release %s has no asset named %s", release.TagName, binaryName)
	}
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(expectedDigest) {
		return fmt.Errorf("release asset %s has no valid SHA-256 digest", binaryName)
	}
	parsedURL, err := url.Parse(downloadURL)
	if err != nil || parsedURL.Scheme != "https" || !trustedDownloadHost(parsedURL) {
		return fmt.Errorf("release asset has an untrusted download URL")
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}
	currentInfo, err := os.Stat(execPath)
	if err != nil {
		return fmt.Errorf("failed to inspect executable: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("User-Agent", "maniplacer/"+utils.Version)
	res, err := updateHTTPClient(5 * time.Minute).Do(req)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", res.StatusCode)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(execPath), ".maniplacer-update-*")
	if err != nil {
		return fmt.Errorf("failed to create update file beside executable: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(tempFile, hash), io.LimitReader(res.Body, maxBinarySize+1))
	if copyErr != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write update file: %w", copyErr)
	}
	if written > maxBinarySize {
		tempFile.Close()
		return fmt.Errorf("download exceeds %d bytes", maxBinarySize)
	}
	if actual := hex.EncodeToString(hash.Sum(nil)); actual != expectedDigest {
		tempFile.Close()
		return fmt.Errorf("SHA-256 verification failed for %s", binaryName)
	}
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to sync update file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close update file: %w", err)
	}
	if err := os.Chmod(tempPath, currentInfo.Mode().Perm()); err != nil {
		return fmt.Errorf("failed to make update executable: %w", err)
	}

	versionOutput, err := exec.CommandContext(ctx, tempPath, "--version").Output()
	if err != nil {
		return fmt.Errorf("downloaded binary failed its version check: %w", err)
	}
	if got, want := normalizeVersion(strings.TrimSpace(string(versionOutput))), normalizeVersion(release.TagName); got != want {
		return fmt.Errorf("downloaded binary reports version %q, expected %q", got, want)
	}

	if err := os.Rename(tempPath, execPath); err != nil {
		return fmt.Errorf("failed to atomically replace executable: %w", err)
	}
	if dir, err := os.Open(filepath.Dir(execPath)); err == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	return nil
}

func releaseBinaryName(goos, goarch string) (string, error) {
	switch goos + "/" + goarch {
	case "linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64":
		return fmt.Sprintf("maniplacer-%s-%s", goos, goarch), nil
	case "windows/amd64":
		return "maniplacer-windows-amd64.exe", nil
	default:
		return "", fmt.Errorf("unsupported platform %s/%s", goos, goarch)
	}
}

var semanticVersionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:[-+].*)?$`)

func normalizeVersion(version string) string {
	return strings.TrimPrefix(strings.TrimSpace(version), "v")
}

func isNewerVersion(current, latest string) (bool, error) {
	latestParts, ok := parseSemanticVersion(latest)
	if !ok {
		return false, fmt.Errorf("latest release has invalid semantic version %q", latest)
	}
	currentParts, ok := parseSemanticVersion(current)
	if !ok {
		return true, nil
	}
	for i := range latestParts {
		if latestParts[i] != currentParts[i] {
			return latestParts[i] > currentParts[i], nil
		}
	}
	return false, nil
}

func parseSemanticVersion(version string) ([3]int, bool) {
	var result [3]int
	matches := semanticVersionPattern.FindStringSubmatch(strings.TrimSpace(version))
	if matches == nil {
		return result, false
	}
	for i := range result {
		part, err := strconv.Atoi(matches[i+1])
		if err != nil {
			return result, false
		}
		result[i] = part
	}
	return result, true
}
