package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectConfigFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expect   ConfigFormat
	}{
		{"json file", "config.json", FormatJSON},
		{"yaml file", "config.yaml", FormatYAML},
		{"yml file", "config.yml", FormatYML},
		{"unknown extension", "config.txt", FormatJSON},
		{"no extension", "config", FormatJSON},
		{"uppercase", "config.JSON", FormatJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectConfigFormat(tt.filename)
			if result != tt.expect {
				t.Errorf("DetectConfigFormat(%q) = %v, want %v", tt.filename, result, tt.expect)
			}
		})
	}
}

func TestConfigLoader_LoadConfig_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	configData := map[string]any{
		"name":      "test-app",
		"namespace": "production",
		"replicas":  3,
	}

	content, _ := json.MarshalIndent(configData, "", "  ")
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := &ConfigLoader{
		FilePath: configPath,
		Format:   FormatJSON,
	}

	result, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if result["name"] != "test-app" {
		t.Errorf("Expected name = 'test-app', got %v", result["name"])
	}

	if result["namespace"] != "production" {
		t.Errorf("Expected namespace = 'production', got %v", result["namespace"])
	}

	if replicas, ok := result["replicas"].(float64); !ok || replicas != 3 {
		t.Errorf("Expected replicas = 3, got %v", result["replicas"])
	}
}

func TestConfigLoader_LoadConfig_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configData := `
name: test-app
namespace: staging
replicas: 5
`
	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := &ConfigLoader{
		FilePath: configPath,
		Format:   FormatYAML,
	}

	result, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if result["name"] != "test-app" {
		t.Errorf("Expected name = 'test-app', got %v", result["name"])
	}

	if result["namespace"] != "staging" {
		t.Errorf("Expected namespace = 'staging', got %v", result["namespace"])
	}

	// YAML parses numbers as interface{}, check with type assertion
	replicas := result["replicas"]
	if replicas == nil {
		t.Errorf("Expected replicas to exist")
	}
}

func TestConfigLoader_LoadConfig_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(configPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := &ConfigLoader{
		FilePath: configPath,
		Format:   FormatJSON,
	}

	_, err := loader.LoadConfig()
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestConfigLoader_LoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("invalid: yaml: :"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := &ConfigLoader{
		FilePath: configPath,
		Format:   FormatYAML,
	}

	_, err := loader.LoadConfig()
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestConfigLoader_LoadConfig_NonExistent(t *testing.T) {
	loader := &ConfigLoader{
		FilePath: "/non/existent/path/config.json",
		Format:   FormatJSON,
	}

	_, err := loader.LoadConfig()
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]any
		wantErr bool
	}{
		{"valid config", map[string]any{"key": "value"}, false},
		{"empty config", map[string]any{}, true},
		{"nil config", nil, true},
		{"multiple keys", map[string]any{"key1": "value1", "key2": "value2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAutoDetectConfigFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Test JSON detection
	jsonPath := filepath.Join(tmpDir, "config.json")
	jsonContent := `{"name": "test"}`
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write JSON file: %v", err)
	}

	format, err := AutoDetectConfigFormat(jsonPath)
	if err != nil {
		t.Fatalf("AutoDetectConfigFormat() error = %v", err)
	}
	if format != FormatJSON {
		t.Errorf("AutoDetectConfigFormat() JSON = %v, want %v", format, FormatJSON)
	}

	// Test YAML detection
	yamlPath := filepath.Join(tmpDir, "config.yaml")
	yamlContent := "name: test"
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	format, err = AutoDetectConfigFormat(yamlPath)
	if err != nil {
		t.Fatalf("AutoDetectConfigFormat() error = %v", err)
	}
	if format != FormatYAML {
		t.Errorf("AutoDetectConfigFormat() YAML = %v, want %v", format, FormatYAML)
	}
}

func TestFindConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	repo := "test-repo"
	repoPath := filepath.Join(tmpDir, repo)

	// Create repo directory
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create repo directory: %v", err)
	}

	t.Run("single config file", func(t *testing.T) {
		// Create only JSON config
		configPath := filepath.Join(repoPath, "config.json")
		if err := os.WriteFile(configPath, []byte(`{"test": "value"}`), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		foundPath, format, err := FindConfigFile(tmpDir, repo, "")
		if err != nil {
			t.Fatalf("FindConfigFile() error = %v", err)
		}
		if foundPath != configPath {
			t.Errorf("FindConfigFile() path = %v, want %v", foundPath, configPath)
		}
		if format != FormatJSON {
			t.Errorf("FindConfigFile() format = %v, want %v", format, FormatJSON)
		}
	})

	t.Run("no config file", func(t *testing.T) {
		// Remove any config files
		os.RemoveAll(repoPath)
		os.MkdirAll(repoPath, 0755)

		_, _, err := FindConfigFile(tmpDir, repo, "")
		if err == nil {
			t.Error("Expected error for no config file, got nil")
		}
	})
}
