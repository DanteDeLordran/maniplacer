package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"text/template"
	"time"

	"github.com/dantedelordran/maniplacer/internal/models"
)

const VERSION = "1.0.0"

func Help() {
	fmt.Println("Usage: maniplacer new -f <path to json>")
}

func Version() {
	fmt.Println(VERSION)
}

func NewManifest() {
	cmd := flag.NewFlagSet("new", flag.ExitOnError)
	file := cmd.String("f", "", "Path to JSON config file")

	cmd.Parse(os.Args[2:])

	if *file == "" {
		fmt.Println("Error: -f flag is required")
		cmd.Usage()
		os.Exit(1)
	}

	config, err := loadJson(*file)

	if err != nil {
		fmt.Println("Error loading config due to ", err)
		os.Exit(1)
	}

	manifest, err := os.ReadFile(filepath.Join("internal", "manifest", "manifest.yml"))

	if err != nil {
		fmt.Println("Error loading yaml due to ", err)
		os.Exit(1)
	}

	yml, err := replaceYaml(manifest, config)

	if err != nil {
		fmt.Println("Error replacing yaml due to ", err)
		os.Exit(1)
	}

	home, err := os.UserHomeDir()

	if err != nil {
		fmt.Println("Error getting HOME dir due to ", err)
		os.Exit(1)
	}

	filename := fmt.Sprintf("manifest-%s-%s.yaml", config.NameSpace, time.Now().Format("20060102-150405"))

	err = os.MkdirAll(filepath.Join(home, "maniplacer"), 0644)

	if err != nil {
		fmt.Println("Error creating dir due to ", err)
		os.Exit(1)
	}

	if err := os.WriteFile(filepath.Join(home, "maniplacer", filename), yml, 0644); err != nil {
		fmt.Printf("Error saving manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Succesfuly created file at: ", filepath.Join(home, "maniplacer", filename))

}

func AutoUpdate() {
	cmd := flag.NewFlagSet("update", flag.ExitOnError)
	check := cmd.Bool("c", false, "Check if there is an update")
	cmd.Parse(os.Args[2:])

	version, err := getLatestVersion()
	if err != nil {
		fmt.Println("Could not get latest version due to: ", err)
		os.Exit(1)
	}

	if *check {
		if version == VERSION {
			fmt.Println("No new version available")
		} else {
			fmt.Println("New version available: ", version)
		}
		return
	}

	if version == VERSION {
		fmt.Println("No new version available")
		return
	}

	fmt.Printf("Updating from %s to %s...\n", VERSION, version)
	if err := downloadAndReplace(version); err != nil {
		fmt.Printf("Update failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Succesfully updated to", version)

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

	var release models.GitHubRelease
	if err := json.NewDecoder(res.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode release info: %w", err)
	}

	// Remove 'v' prefix if present (e.g., "v1.0.0" -> "1.0.0")
	version := release.TagName
	if len(version) > 0 && version[0] == 'v' {
		version = version[1:]
	}

	return version, nil
}

func downloadAndReplace(version string) error {
	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Determine the binary name based on OS
	goos := runtime.GOOS
	arch := runtime.GOARCH

	// Construct the expected binary name
	binaryName := fmt.Sprintf("maniplacer-%s-%s", goos, arch)
	if goos == "windows" {
		binaryName += ".exe"
	}

	// Get the download URL
	downloadURL, err := getDownloadURL(version, binaryName)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	// Download the new binary
	fmt.Printf("Downloading %s...\n", downloadURL)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "maniplacer-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Copy the downloaded content to temp file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tempFile.Close()

	// Make the temp file executable
	if err := os.Chmod(tempFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make temp file executable: %w", err)
	}

	// Create backup of current binary
	backupPath := execPath + ".backup"
	if err := copyFile(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Replace the current binary
	if err := copyFile(tempFile.Name(), execPath); err != nil {
		// Restore from backup if replacement fails
		copyFile(backupPath, execPath)
		os.Remove(backupPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Clean up backup
	os.Remove(backupPath)

	return nil
}

func getDownloadURL(version, binaryName string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/dantedelordran/maniplacer/releases/tags/%s", version)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release models.GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode release info: %w", err)
	}

	// Find the matching asset
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

	// Copy file permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

func loadJson(path string) (*models.ManifestConfig, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	var config models.ManifestConfig
	err = json.Unmarshal(data, &config)
	return &config, nil
}

func replaceYaml(templateContent []byte, config *models.ManifestConfig) ([]byte, error) {

	tmpl, err := template.New("manifest").Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("template parsing failed: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return nil, fmt.Errorf("template execution failed: %w", err)
	}

	return buf.Bytes(), nil
}
