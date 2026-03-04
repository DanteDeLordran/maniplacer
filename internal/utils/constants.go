package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// Constants for default values
const (
	DefaultNamespace = "default"
	DefaultPort      = "8000"
	ConfigFileName   = "config"
	ManiplacerMarker = ".maniplacer"
)

// Supported config formats
const (
	FormatJSON = "json"
	FormatYAML = "yaml"
	FormatYML  = "yml"
)

// File permissions
const (
	DirPermission  = 0755
	FilePermission = 0644
)

// Kubernetes naming conventions
// DNS-1123 label: must consist of lower case alphanumeric characters or '-',
// and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc')
var dns1123LabelRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// DNS-1123 subdomain: lowercase alphanumeric characters, '-' or '.',
// must start and end with alphanumeric character
var dns1123SubdomainRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

// ValidateRepoName validates repository name according to K8s naming conventions
func ValidateRepoName(name string) error {
	if name == "" {
		return fmt.Errorf("repository name cannot be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("repository name must be 63 characters or less")
	}
	if !dns1123LabelRegex.MatchString(name) {
		return fmt.Errorf("repository name must consist of lowercase alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name', '123-abc')")
	}
	return nil
}

// ValidateNamespace validates namespace name according to K8s naming conventions
func ValidateNamespace(name string) error {
	if name == "" {
		return fmt.Errorf("namespace name cannot be empty")
	}
	if name == "kube-system" || name == "kube-public" || name == "kube-node-lease" {
		return fmt.Errorf("namespace '%s' is reserved by Kubernetes", name)
	}
	if len(name) > 63 {
		return fmt.Errorf("namespace name must be 63 characters or less")
	}
	if !dns1123LabelRegex.MatchString(name) {
		return fmt.Errorf("namespace name must consist of lowercase alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name', '123-abc')")
	}
	return nil
}

// ValidateProjectName validates project name (allows dots for subdomains)
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if len(name) > 253 {
		return fmt.Errorf("project name must be 253 characters or less")
	}
	if !dns1123SubdomainRegex.MatchString(name) {
		return fmt.Errorf("project name must consist of lowercase alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
	}
	return nil
}

// SanitizeName sanitizes a name to be K8s-compliant
func SanitizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	// Replace invalid characters with '-'
	name = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(name, "-")
	// Remove leading/trailing dashes
	name = strings.Trim(name, "-")
	// Limit length
	if len(name) > 63 {
		name = name[:63]
		name = strings.TrimRight(name, "-")
	}
	return name
}

// IsPathTraversal checks if a path contains path traversal attempts
func IsPathTraversal(path string) bool {
	// Check for common path traversal patterns
	dangerousPatterns := []string{
		"..",
		"./",
		"/..",
		"../",
		"\\..",
		"..\\",
	}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// ValidateSafePath validates that a path doesn't contain path traversal
func ValidateSafePath(path string) error {
	if IsPathTraversal(path) {
		return fmt.Errorf("path '%s' contains invalid path traversal sequence", path)
	}
	return nil
}
