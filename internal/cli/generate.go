package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/dantedelordran/maniplacer/internal/templates"
	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type ConfigFormat string

const (
	FormatJSON ConfigFormat = "json"
	FormatYAML ConfigFormat = "yaml"
	FormatYML  ConfigFormat = "yml"
)

type ConfigLoader struct {
	FilePath string
	Format   ConfigFormat
}

func (cl *ConfigLoader) LoadConfig() (map[string]any, error) {
	content, err := os.ReadFile(cl.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", cl.FilePath, err)
	}

	var config map[string]any

	switch cl.Format {
	case FormatJSON:
		if err := json.Unmarshal(content, &config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config file '%s': %w", cl.FilePath, err)
		}
	case FormatYAML, FormatYML:
		if err := yaml.Unmarshal(content, &config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config file '%s': %w", cl.FilePath, err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format: %s", cl.Format)
	}

	return config, nil
}

func DetectConfigFormat(filePath string) ConfigFormat {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return FormatJSON
	case ".yaml":
		return FormatYAML
	case ".yml":
		return FormatYML
	default:
		return FormatJSON // Default fallback
	}
}

type ConfigCandidate struct {
	Path     string
	Filename string
	Format   ConfigFormat
}

func FindConfigFile(baseDir, repo string, preferredFormat ConfigFormat) (string, ConfigFormat, error) {
	configDir := filepath.Join(baseDir, repo)

	// If a specific format is requested, try that first
	if preferredFormat != "" {
		configPath := filepath.Join(configDir, fmt.Sprintf("config.%s", preferredFormat))
		if _, err := os.Stat(configPath); err == nil {
			return configPath, preferredFormat, nil
		}
		// If preferred format not found, continue with auto-detection
		fmt.Printf("Warning: Preferred config format '%s' not found, auto-detecting...\n", preferredFormat)
	}

	// Define all possible config file candidates
	candidates := []struct {
		filename string
		format   ConfigFormat
	}{
		{"config.json", FormatJSON},
		{"config.yaml", FormatYAML},
		{"config.yml", FormatYML},
	}

	// Find all existing config files
	var foundCandidates []ConfigCandidate
	for _, candidate := range candidates {
		configPath := filepath.Join(configDir, candidate.filename)
		if _, err := os.Stat(configPath); err == nil {
			foundCandidates = append(foundCandidates, ConfigCandidate{
				Path:     configPath,
				Filename: candidate.filename,
				Format:   candidate.format,
			})
		}
	}

	// Handle different scenarios
	switch len(foundCandidates) {
	case 0:
		return "", "", fmt.Errorf("no configuration file found in %s (tried: config.json, config.yaml, config.yml)", configDir)

	case 1:
		// Only one config file found, use it
		candidate := foundCandidates[0]
		return candidate.Path, candidate.Format, nil

	default:
		// Multiple config files found, ask user to choose
		return promptForConfigChoice(foundCandidates)
	}
}

// promptForConfigChoice asks the user to choose between multiple config files
func promptForConfigChoice(candidates []ConfigCandidate) (string, ConfigFormat, error) {
	fmt.Printf("\nMultiple configuration files found:\n")
	for i, candidate := range candidates {
		fmt.Printf("  %d) %s\n", i+1, candidate.Filename)
	}

	fmt.Printf("\nPlease choose which config file to use (1-%d): ", len(candidates))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("failed to read user input: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil {
		return "", "", fmt.Errorf("invalid input '%s': please enter a number", input)
	}

	if choice < 1 || choice > len(candidates) {
		return "", "", fmt.Errorf("invalid choice %d: please choose between 1 and %d", choice, len(candidates))
	}

	selected := candidates[choice-1]
	fmt.Printf("Selected: %s\n", selected.Filename)

	return selected.Path, selected.Format, nil
}

// AutoDetectConfigFormat attempts to detect format by parsing the file content
func AutoDetectConfigFormat(filePath string) (ConfigFormat, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file for format detection: %w", err)
	}

	// Try JSON first
	var jsonTest map[string]any
	if err := json.Unmarshal(content, &jsonTest); err == nil {
		return FormatJSON, nil
	}

	// Try YAML
	var yamlTest map[string]any
	if err := yaml.Unmarshal(content, &yamlTest); err == nil {
		return FormatYAML, nil
	}

	// If both fail, default to extension-based detection
	return DetectConfigFormat(filePath), nil
}
func ValidateConfig(config map[string]any) error {
	if len(config) == 0 {
		return fmt.Errorf("configuration file is empty or contains no valid data")
	}
	return nil
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates the manifest given a config file and at least one template",
	Long: `The generate command renders Kubernetes manifests by combining your project's templates with values provided in a configuration file.

It scans the 'templates/<namespace>/' directory of your Maniplacer project, applies the values from the configuration file, and writes the rendered manifests into 'manifests/<namespace>/<timestamp>/'. 
Each run is timestamped, ensuring outputs from previous runs are preserved instead of being overwritten.

You can customize the input config format with the --format (or -f) flag, select a template namespace with the --namespace (or -n) flag, and specify the target repository with the --repo (or -r) flag.

Supported config formats: JSON (.json), YAML (.yaml, .yml)

Typical workflow:
1. Define your application values in a config file (config.json, config.yaml, or config.yml).
2. Create or edit Kubernetes resource templates under 'templates/<namespace>/'.
3. Run 'maniplacer generate' to render manifests into a dedicated timestamped output folder.

Example usage:
  maniplacer generate
  maniplacer generate -n staging
  maniplacer generate -f yaml -n production -r myrepo
  maniplacer generate -c /path/to/custom-config.json
  maniplacer generate -c custom.yaml -f yaml

Notes:
- The current directory must be a valid Maniplacer project (contain a '.maniplacer' file).
- The specified namespace must exist under the 'templates' directory.
- Each run creates a unique timestamped output folder for safe, repeatable generation.`,
	Args: cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		if !utils.IsValidProject() {
			fmt.Printf("Error: Current directory is not a valid Maniplacer project\n")
			os.Exit(1)
		}

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			fmt.Printf("Warning: Could not parse namespace flag, using default\n")
			namespace = "default"
		}

		formatFlag, err := cmd.Flags().GetString("format")
		if err != nil {
			fmt.Printf("Warning: Could not parse format flag, using auto-detection\n")
			formatFlag = ""
		}

		customConfigPath, err := cmd.Flags().GetString("config")
		if err != nil {
			fmt.Printf("Warning: Could not parse config flag\n")
			customConfigPath = ""
		}

		repo, err := cmd.Flags().GetString("repo")
		if err != nil {
			fmt.Printf("Error: Could not get repo flag: %s\n", err)
			os.Exit(1)
		}

		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error: Could not get current directory: %s\n", err)
			os.Exit(1)
		}

		templateDir := filepath.Join(currentDir, repo, "templates", namespace)
		if _, err := os.Stat(templateDir); err != nil {
			fmt.Printf("Error: Template directory '%s' not found: %s\n", templateDir, err)
			os.Exit(1)
		}

		files, err := os.ReadDir(templateDir)
		if err != nil {
			fmt.Printf("Error: Could not read template directory: %s\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Printf("Error: Template namespace '%s' is empty\n", namespace)
			os.Exit(1)
		}

		// Find and load configuration file
		var configPath string
		var detectedFormat ConfigFormat

		if customConfigPath != "" {
			// Use custom config file path
			if !filepath.IsAbs(customConfigPath) {
				customConfigPath = filepath.Join(currentDir, repo, customConfigPath)
			}

			if _, err := os.Stat(customConfigPath); err != nil {
				fmt.Printf("Error: Custom config file not found: %s\n", customConfigPath)
				os.Exit(1)
			}

			configPath = customConfigPath

			// Determine format for custom config file
			if formatFlag != "" {
				// Format explicitly specified
				detectedFormat = ConfigFormat(strings.ToLower(formatFlag))
				// Validate the format flag
				switch detectedFormat {
				case FormatJSON, FormatYAML, FormatYML:
					// Valid format
				default:
					fmt.Printf("Error: Unsupported format '%s'. Supported formats: json, yaml, yml\n", formatFlag)
					os.Exit(1)
				}
			} else {
				// Auto-detect format for custom file
				detectedFormat, err = AutoDetectConfigFormat(customConfigPath)
				if err != nil {
					fmt.Printf("Error: Could not detect format for custom config file: %s\n", err)
					os.Exit(1)
				}
			}
		} else {
			// Use standard config file detection
			var preferredFormat ConfigFormat
			if formatFlag != "" {
				preferredFormat = ConfigFormat(strings.ToLower(formatFlag))
				// Validate the format flag
				switch preferredFormat {
				case FormatJSON, FormatYAML, FormatYML:
					// Valid format
				default:
					fmt.Printf("Error: Unsupported format '%s'. Supported formats: json, yaml, yml\n", formatFlag)
					os.Exit(1)
				}
			}

			configPath, detectedFormat, err = FindConfigFile(currentDir, repo, preferredFormat)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("Using %s config file: %s\n", strings.ToUpper(string(detectedFormat)), configPath)

		loader := &ConfigLoader{
			FilePath: configPath,
			Format:   detectedFormat,
		}

		config, err := loader.LoadConfig()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		if err := ValidateConfig(config); err != nil {
			fmt.Printf("Error: Configuration validation failed: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully loaded configuration with %d top-level keys\n", len(config))

		// Generate output directory with timestamp
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		outputDir := filepath.Join(currentDir, repo, "manifests", namespace, timestamp)

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error: Could not create output directory '%s': %s\n", outputDir, err)
			os.Exit(1)
		}

		fmt.Printf("Output directory: %s\n", outputDir)

		// Process templates
		successCount := 0
		errorCount := 0

		for _, file := range files {
			if file.IsDir() {
				continue // Skip directories
			}

			templatePath := filepath.Join(templateDir, file.Name())

			if err := processTemplate(templatePath, outputDir, file.Name(), config); err != nil {
				fmt.Printf("Warning: Failed to process template '%s': %s\n", file.Name(), err)
				errorCount++
			} else {
				fmt.Printf("Generated: %s\n", filepath.Join(outputDir, file.Name()))
				successCount++
			}
		}

		fmt.Printf("\nGeneration complete: %d successful, %d errors\n", successCount, errorCount)

		if errorCount > 0 {
			os.Exit(1)
		}
	},
}

// processTemplate handles the rendering of a single template file
func processTemplate(templatePath, outputDir, filename string, config map[string]any) error {
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("could not read template file: %w", err)
	}

	templ, err := template.New(filename).Funcs(templates.ManiplacerFuncs).Parse(string(content))
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	outputPath := filepath.Join(outputDir, filename)
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("could not create output file: %w", err)
	}
	defer f.Close()

	if err := templ.Execute(f, config); err != nil {
		// Clean up the partially written file on error
		f.Close()
		os.Remove(outputPath)
		return fmt.Errorf("could not execute template: %w", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("format", "f", "", "Config file format (json, yaml, yml). If not specified, auto-detects from available files.")
	generateCmd.Flags().StringP("namespace", "n", "default", "Namespace for template to be generated")
	generateCmd.Flags().StringP("repo", "r", "", "Repository name")
	generateCmd.Flags().StringP("config", "c", "", "Custom path to config file (overrides default config file detection)")
}
