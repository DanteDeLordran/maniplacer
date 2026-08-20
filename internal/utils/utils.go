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

	return os.WriteFile(filepath.Join(path, ManiplacerMarker), data, FilePermission)
}

func IsValidProject() bool {
	data, err := os.ReadFile(ManiplacerMarker)
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
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Could not read input due to: %s\n", err)
		os.Exit(1)
	}
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// starterConfig is the seed content written into a new repo's config file. An
// empty file is not valid JSON or a usable YAML mapping, so 'generate' would
// fail on it straight away; this gives a parseable document to build on.
const (
	starterConfigJSON = `{
  "name": "my-app",
  "namespace": "default",
  "image": "nginx:1.27-alpine",
  "replicas": 2,
  "containerPort": 8080,
  "servicePort": 80,
  "configValue": "development",
  "secretPlaceholder": "replace-at-render-time",
  "minReplicas": 2,
  "maxReplicas": 5,
  "cpuUtilization": 75,
  "memoryUtilization": 80,
  "gatewayName": "main-gateway",
  "gatewayNamespace": "default",
  "hostname": "my-app.example.com",
  "pathPrefix": "/",
  "healthPath": "/healthz"
}
`
	starterConfigYAML = `name: my-app
namespace: default
image: nginx:1.27-alpine
replicas: 2
containerPort: 8080
servicePort: 80
configValue: development
secretPlaceholder: replace-at-render-time
minReplicas: 2
maxReplicas: 5
cpuUtilization: 75
memoryUtilization: 80
gatewayName: main-gateway
gatewayNamespace: default
hostname: my-app.example.com
pathPrefix: /
healthPath: /healthz
`
)

// CreateConfigFile writes a starter config file of the given type (json, yaml or
// yml; anything else falls back to json) into path.
func CreateConfigFile(path string, filetype string) error {
	filetype = strings.ToLower(strings.TrimPrefix(filetype, "."))

	content := starterConfigJSON
	switch filetype {
	case "yaml", "yml":
		content = starterConfigYAML
	default:
		filetype = "json"
	}

	configPath := filepath.Join(path, fmt.Sprintf("%s.%s", ConfigFileName, filetype))
	if err := os.WriteFile(configPath, []byte(content), FilePermission); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	return nil
}
