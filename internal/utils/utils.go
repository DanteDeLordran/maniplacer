package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func ConfirmMessage(message string) bool {
	fmt.Printf("%s [y/N]: ", message)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}
