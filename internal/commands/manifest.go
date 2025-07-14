package commands

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/dantedelordran/maniplacer/internal/models"
)

//go:embed manifest/manifest.yml
var embedManifest embed.FS

func NewManifest() {
	cmd := flag.NewFlagSet("new", flag.ExitOnError)
	file := cmd.String("f", "", "Path to JSON config file")
	target := cmd.String("t", "", "Path for the manifest to be stored")
	cmd.Parse(os.Args[2:])

	if *file == "" {
		fmt.Println("Error: -f flag is required")
		cmd.Usage()
		os.Exit(1)
	}

	config, err := loadJson(*file)
	if err != nil {
		fmt.Println("Error loading config due to", err)
		os.Exit(1)
	}

	//manifest, err := os.ReadFile(filepath.Join("internal", "manifest", "manifest.yml"))
	manifest, err := embedManifest.ReadFile("manifest/manifest.yml")
	if err != nil {
		fmt.Println("Error loading embeded yaml due to", err)
		os.Exit(1)
	}

	yml, err := replaceYaml(manifest, config)
	if err != nil {
		fmt.Println("Error replacing yaml due to", err)
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting HOME dir due to", err)
		os.Exit(1)
	}

	filename := fmt.Sprintf("%s-%s-%s.yaml", config.Name, config.NameSpace, time.Now().Format("20060102-150405"))

	if *target == "" {
		*target = "/maniplacer"
	} else {
		if strings.HasPrefix(*target, home) {
			*target = strings.Replace(*target, home, "", -1)
		}
	}

	outputDir := filepath.Join(home, *target)

	if err := os.MkdirAll(outputDir, 0744); err != nil {
		fmt.Println("Error creating dir due to", err)
		os.Exit(1)
	}

	outputPath := filepath.Join(outputDir, filename)
	if err := os.WriteFile(outputPath, yml, 0644); err != nil {
		fmt.Printf("Error saving manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully created file at:", outputPath)
}

func loadJson(path string) (*models.ManifestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	var config models.ManifestConfig
	err = json.Unmarshal(data, &config)
	return &config, err
}

func replaceYaml(templateContent []byte, config *models.ManifestConfig) ([]byte, error) {
	tmpl, err := template.New("manifest").Funcs(template.FuncMap{"b64enc": func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}}).Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("template parsing failed: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return nil, fmt.Errorf("template execution failed: %w", err)
	}

	return buf.Bytes(), nil
}
