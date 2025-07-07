package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/dantedelordran/maniplacer/internal/models"
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
