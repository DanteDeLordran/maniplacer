package cli

import (
	"bufio"
	"context"
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
  maniplacer generate --dry-run

Notes:
- The current directory must be a valid Maniplacer project (contain a '.maniplacer' file).
- The specified namespace must exist under the 'templates' directory.
- Each run creates a unique timestamped output folder for safe, repeatable generation.
- Use --dry-run to preview without writing files.`,
	Args: cobra.MaximumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.LoggerFromContext(cmd.Context())

		if !utils.IsValidProject() {
			return fmt.Errorf("current directory is not a valid Maniplacer project")
		}

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logger.Debug("could not parse namespace flag, using default", "error", err)
			namespace = utils.DefaultNamespace
		}

		// Validate namespace
		if err := utils.ValidateNamespace(namespace); err != nil {
			return fmt.Errorf("invalid namespace: %w", err)
		}

		formatFlag, err := cmd.Flags().GetString("format")
		if err != nil {
			logger.Debug("could not parse format flag, using auto-detection", "error", err)
			formatFlag = ""
		}

		customConfigPath, err := cmd.Flags().GetString("config")
		if err != nil {
			logger.Debug("could not parse config flag", "error", err)
			customConfigPath = ""
		}

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			logger.Debug("could not parse dry-run flag", "error", err)
			dryRun = false
		}

		repo, err := cmd.Flags().GetString("repo")
		if err != nil {
			return fmt.Errorf("could not get repo flag: %w", err)
		}

		if repo == "" {
			return fmt.Errorf("repository name is required (use --repo flag)")
		}

		// Validate repo name and check for path traversal
		if err := utils.ValidateRepoName(repo); err != nil {
			return fmt.Errorf("invalid repository name: %w", err)
		}
		if err := utils.ValidateSafePath(repo); err != nil {
			return err
		}

		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get current directory: %w", err)
		}

		templateDir := filepath.Join(currentDir, repo, "templates", namespace)
		if _, err := os.Stat(templateDir); err != nil {
			return fmt.Errorf("template directory '%s' not found: %w", templateDir, err)
		}

		files, err := os.ReadDir(templateDir)
		if err != nil {
			return fmt.Errorf("could not read template directory: %w", err)
		}

		if len(files) == 0 {
			return fmt.Errorf("template namespace '%s' is empty", namespace)
		}

		// Find and load configuration file
		var configPath string
		var detectedFormat ConfigFormat

		if customConfigPath != "" {
			// Use custom config file path
			if !filepath.IsAbs(customConfigPath) {
				customConfigPath = filepath.Join(currentDir, repo, customConfigPath)
			}

			// Validate custom config path for path traversal
			if err := utils.ValidateSafePath(customConfigPath); err != nil {
				return err
			}

			if _, err := os.Stat(customConfigPath); err != nil {
				return fmt.Errorf("custom config file not found: %s", customConfigPath)
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
					return fmt.Errorf("unsupported format '%s'. Supported formats: json, yaml, yml", formatFlag)
				}
			} else {
				// Auto-detect format for custom file
				detectedFormat, err = AutoDetectConfigFormat(customConfigPath)
				if err != nil {
					return fmt.Errorf("could not detect format for custom config file: %w", err)
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
					return fmt.Errorf("unsupported format '%s'. Supported formats: json, yaml, yml", formatFlag)
				}
			}

			configPath, detectedFormat, err = FindConfigFile(currentDir, repo, preferredFormat)
			if err != nil {
				return err
			}
		}

		logger.Info("using config file", "format", strings.ToUpper(string(detectedFormat)), "path", configPath)
		fmt.Printf("Using %s config file: %s\n", strings.ToUpper(string(detectedFormat)), configPath)

		loader := &ConfigLoader{
			FilePath: configPath,
			Format:   detectedFormat,
		}

		config, err := loader.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := ValidateConfig(config); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}

		logger.Info("configuration loaded successfully", "keys", len(config))
		fmt.Printf("Successfully loaded configuration with %d top-level keys\n", len(config))

		// Generate output directory with timestamp
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		outputDir := filepath.Join(currentDir, repo, "manifests", namespace, timestamp)

		if !dryRun {
			if err := os.MkdirAll(outputDir, utils.DirPermission); err != nil {
				return fmt.Errorf("could not create output directory '%s': %w", outputDir, err)
			}
			logger.Info("output directory created", "path", outputDir)
			fmt.Printf("Output directory: %s\n", outputDir)
		} else {
			logger.Info("dry-run mode enabled, no files will be written")
			fmt.Printf("Dry-run mode: no files will be written\n")
		}

		// Process templates
		successCount := 0
		errorCount := 0

		for _, file := range files {
			if file.IsDir() {
				continue // Skip directories
			}

			templatePath := filepath.Join(templateDir, file.Name())

			if err := processTemplate(cmd.Context(), templatePath, outputDir, file.Name(), config, dryRun); err != nil {
				logger.Warn("failed to process template", "file", file.Name(), "error", err)
				fmt.Printf("Warning: Failed to process template '%s': %s\n", file.Name(), err)
				errorCount++
			} else {
				if dryRun {
					fmt.Printf("Would generate: %s\n", file.Name())
				} else {
					logger.Info("manifest generated", "file", file.Name())
					fmt.Printf("Generated: %s\n", filepath.Join(outputDir, file.Name()))
				}
				successCount++
			}
		}

		logger.Info("generation complete", "successful", successCount, "errors", errorCount)
		fmt.Printf("\nGeneration complete: %d successful, %d errors\n", successCount, errorCount)

		if errorCount > 0 {
			return fmt.Errorf("generation completed with %d errors", errorCount)
		}

		return nil
	},
}

// processTemplate handles the rendering of a single template file
func processTemplate(ctx context.Context, templatePath, outputDir, filename string, config map[string]any, dryRun bool) error {
	logger := utils.LoggerFromContext(ctx)

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("could not read template file: %w", err)
	}

	templ, err := template.New(filename).Funcs(templates.ManiplacerFuncs).Parse(string(content))
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	if dryRun {
		// In dry-run mode, just validate the template without writing
		var output strings.Builder
		if err := templ.Execute(&output, config); err != nil {
			return fmt.Errorf("could not execute template: %w", err)
		}
		logger.Debug("template validated successfully", "file", filename)
		return nil
	}

	outputPath := filepath.Join(outputDir, filename)
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("could not create output file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logger.Warn("failed to close output file", "file", outputPath, "error", closeErr)
		}
	}()

	if err := templ.Execute(f, config); err != nil {
		// Clean up the partially written file on error
		f.Close()
		if removeErr := os.Remove(outputPath); removeErr != nil {
			logger.Warn("failed to remove partial output file", "file", outputPath, "error", removeErr)
		}
		return fmt.Errorf("could not execute template: %w", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("format", "f", "", "Config file format (json, yaml, yml). If not specified, auto-detects from available files.")
	generateCmd.Flags().StringP("namespace", "n", utils.DefaultNamespace, "Namespace for template to be generated")
	generateCmd.Flags().StringP("repo", "r", "", "Repository name")
	generateCmd.Flags().StringP("config", "c", "", "Custom path to config file (overrides default config file detection)")
	generateCmd.Flags().Bool("dry-run", false, "Preview generation without writing files")
}
