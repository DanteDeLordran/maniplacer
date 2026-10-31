package cli

import (
	"bufio"
	"bytes"
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

// versionLayout names generated versions. Every version sorts chronologically by
// name, and time.Parse with this layout also accepts the fractional seconds that
// generate appends to keep rapid runs apart.
const versionLayout = "2006-01-02_15-04-05"

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

// resolveConfigNamespace sets the .namespace template value to the target, so
// rendered manifests agree with the directories they are read from and written
// to. An explicit --namespace is a deliberate override; a config that disagrees
// with the default target is an error rather than a silent redirect.
func resolveConfigNamespace(config map[string]any, namespace string, explicit bool) error {
	if configured, ok := config["namespace"]; ok && !explicit && configured != namespace {
		return fmt.Errorf("config namespace '%v' does not match target namespace '%s': pass --namespace to choose one", configured, namespace)
	}

	config["namespace"] = namespace
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
  maniplacer generate -r myrepo
  maniplacer generate -n staging -r myrepo
  maniplacer generate -f yaml -n production -r myrepo
  maniplacer generate -c /path/to/custom-config.json -r myrepo
  maniplacer generate -c custom.yaml -f yaml -r myrepo
  maniplacer generate --dry-run -r myrepo
  maniplacer generate --strict=false -r myrepo

Notes:
- The current directory must be a valid Maniplacer project (contain a '.maniplacer' file).
- The specified namespace must exist under the 'templates' directory.
- Each run is published to a timestamped output folder only after every template succeeds, so a failed run leaves the previous version as the latest.
- --namespace also sets the .namespace template value, overriding the config file. Without the flag, a config namespace other than the default is an error.
- Use --dry-run to validate without writing files.
- Generation is strict by default: a template referencing a key that is absent from the config is an error, every rendered YAML document must be valid, and none may set a metadata.namespace other than the target. Pass --strict=false to fall back to the lenient behaviour.`,
	Args: cobra.MaximumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.LoggerFromContext(cmd.Context())

		target, err := resolveTarget(cmd)
		if err != nil {
			return err
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

		strict, err := cmd.Flags().GetBool("strict")
		if err != nil {
			logger.Debug("could not parse strict flag, defaulting to strict", "error", err)
			strict = true
		}

		templateDir := target.TemplatesDir()
		if _, err := os.Stat(templateDir); err != nil {
			return fmt.Errorf("template directory '%s' not found: %w", templateDir, err)
		}

		files, err := os.ReadDir(templateDir)
		if err != nil {
			return fmt.Errorf("could not read template directory: %w", err)
		}

		if len(files) == 0 {
			return fmt.Errorf("template namespace '%s' is empty", target.Namespace)
		}

		// Find and load configuration file
		var configPath string
		var detectedFormat ConfigFormat

		if customConfigPath != "" {
			// Use custom config file path
			if !filepath.IsAbs(customConfigPath) {
				customConfigPath = filepath.Join(target.RepoPath(), customConfigPath)
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

			configPath, detectedFormat, err = FindConfigFile(target.Root, target.Repo, preferredFormat)
			if err != nil {
				return err
			}
		}

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

		fmt.Printf("Successfully loaded configuration with %d top-level keys\n", len(config))

		if err := resolveConfigNamespace(config, target.Namespace, cmd.Flags().Changed("namespace")); err != nil {
			return err
		}

		// Render into a staging directory that sits outside the namespace folders
		// searched by apply, list, and prune, then publish it in a single rename.
		version := time.Now().Format(versionLayout + ".000000000")
		outputDir := filepath.Join(target.ManifestsDir(), version)
		stagingDir := filepath.Join(filepath.Dir(target.ManifestsDir()), ".staging-"+version)

		if !dryRun {
			if err := os.MkdirAll(target.ManifestsDir(), utils.DirPermission); err != nil {
				return fmt.Errorf("could not create manifests directory '%s': %w", target.ManifestsDir(), err)
			}
			if err := os.Mkdir(stagingDir, utils.DirPermission); err != nil {
				return fmt.Errorf("could not create staging directory '%s': %w", stagingDir, err)
			}
			defer os.RemoveAll(stagingDir)
		} else {
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

			if err := processTemplate(cmd.Context(), templatePath, stagingDir, file.Name(), config, target.Namespace, dryRun, strict); err != nil {
				logger.Debug("failed to process template", "file", file.Name(), "error", err)
				fmt.Printf("Warning: Failed to process template '%s': %s\n", file.Name(), err)
				errorCount++
			} else {
				if dryRun {
					fmt.Printf("Would generate: %s\n", file.Name())
				} else {
					fmt.Printf("Generated: %s\n", file.Name())
				}
				successCount++
			}
		}

		fmt.Printf("\nGeneration complete: %d successful, %d errors\n", successCount, errorCount)

		if errorCount > 0 {
			return fmt.Errorf("generation completed with %d errors, no version was published", errorCount)
		}

		if dryRun {
			return nil
		}

		if err := os.Rename(stagingDir, outputDir); err != nil {
			return fmt.Errorf("could not publish version '%s': %w", version, err)
		}
		fmt.Printf("Output directory: %s\n", outputDir)

		return nil
	},
}

// processTemplate renders a single template into memory, validates the result,
// and only then writes it out. Rendering before writing means a failed template
// can never leave a partial manifest behind.
//
// In strict mode a template that references a key missing from the config is an
// error instead of silently rendering "<no value>", every rendered document must
// parse as YAML, and none may set a metadata.namespace other than namespace.
func processTemplate(ctx context.Context, templatePath, outputDir, filename string, config map[string]any, namespace string, dryRun, strict bool) error {
	logger := utils.LoggerFromContext(ctx)

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("could not read template file: %w", err)
	}

	templ := template.New(filename).Funcs(templates.ManiplacerFuncs)
	if strict {
		templ = templ.Option("missingkey=error")
	}

	templ, err = templ.Parse(string(content))
	if err != nil {
		return fmt.Errorf("could not parse template: %w", err)
	}

	var rendered bytes.Buffer
	if err := templ.Execute(&rendered, config); err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	if strict {
		objects, err := decodeManifests(rendered.Bytes())
		if err != nil {
			return fmt.Errorf("rendered output is not valid YAML: %w", err)
		}
		for _, obj := range objects {
			if err := checkManifestNamespace(obj, namespace); err != nil {
				return err
			}
		}
	}

	if dryRun {
		logger.Debug("template validated successfully", "file", filename)
		return nil
	}

	outputPath := filepath.Join(outputDir, filename)
	if err := os.WriteFile(outputPath, rendered.Bytes(), utils.FilePermission); err != nil {
		return fmt.Errorf("could not write output file: %w", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(generateCmd)
	addRepoNamespaceFlags(generateCmd, "Namespace for template to be generated")
	generateCmd.Flags().StringP("format", "f", "", "Config file format (json, yaml, yml). If not specified, auto-detects from available files.")
	generateCmd.Flags().StringP("config", "c", "", "Custom path to config file (overrides default config file detection)")
	generateCmd.Flags().Bool("dry-run", false, "Validate generation without writing files")
	generateCmd.Flags().Bool("strict", true, "Fail on config keys missing from the config file and on output that is not valid YAML")
}
