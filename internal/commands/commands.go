package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dantedelordran/maniplacer/internal/models"
	"gopkg.in/yaml.v3"
)

func Help() {
	fmt.Println("Usage: maniplacer new -f <path to json>")
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

	manifest, err := loadYaml(filepath.Join("internal", "manifest", "manifest.yml"))

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

	filename := fmt.Sprintf("manifest-changes-%s.yaml", time.Now().Format("20060102-150405"))

	err = os.MkdirAll(filepath.Join(home, "maniplacer", filename), 0700)

	if err != nil {
		fmt.Println("Error creating dir due to ", err)
		os.Exit(1)
	}

	fmt.Println("Succesfuly created file at: ", filename)
	fmt.Println(yml)

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

func loadYaml(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	var manifest map[string]any
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return manifest, nil
}

func replaceYaml(manifest map[string]any, config *models.ManifestConfig) (map[string]any, error) {
	return nil, nil
}
