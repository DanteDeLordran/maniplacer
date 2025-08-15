package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ManiplacerProject struct {
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

func CreateManiplacerProject(path string) error {
	config := ManiplacerProject{
		Version:     Version,
		Author:      "Your name",
		Description: "Some nice description",
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(filepath.Join(path, ".maniplacer"), data, 0644)
}

func IsValidProject() bool {
	data, err := os.ReadFile(".maniplacer")
	if err != nil {
		return false
	}

	var cfg ManiplacerProject
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false
	}

	return true
}
