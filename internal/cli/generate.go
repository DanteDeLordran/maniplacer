package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/dantedelordran/maniplacer/internal/templates"
	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates the manifest given a config.json file and at least one template",
	Long: `The generate command renders Kubernetes manifests by combining your project’s templates with values provided in a JSON configuration file.

It scans the 'templates/<namespace>/' directory of your Maniplacer project, applies the values from the configuration file (default: config.json), and writes the rendered manifests into 'manifests/<namespace>/<timestamp>/'. 
Each run is timestamped, ensuring outputs from previous runs are preserved instead of being overwritten.

You can customize the input config file with the --file (or -f) flag, select a template namespace with the --namespace (or -n) flag, and specify the target repository with the --repo (or -r) flag.

Typical workflow:
1. Define your application values in a config.json file.
2. Create or edit Kubernetes resource templates under 'templates/<namespace>/'.
3. Run 'maniplacer generate' to render manifests into a dedicated timestamped output folder.

Example usage:
  maniplacer generate
  maniplacer generate -n staging
  maniplacer generate -f custom.json -n production -r myrepo

Notes:
- The current directory must be a valid Maniplacer project (contain a '.maniplacer' file).
- The specified namespace must exist under the 'templates' directory.
- Each run creates a unique timestamped output folder for safe, repeatable generation.`,
	Args: cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		if !utils.IsValidProject() {
			fmt.Printf("Current directory is not a valid Maniplacer project\n")
			os.Exit(1)
		}

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			fmt.Printf("Could not parse namespace flag, using default\n")
			namespace = "default"
		}

		format, err := cmd.Flags().GetString("format")
		if err != nil {
			fmt.Printf("Could not parse file flag, using default\n")
		}

		repo, err := cmd.Flags().GetString("repo")
		if err != nil {
			fmt.Printf("Could not get repo flag due to %s\n", err)
			os.Exit(1)
		}

		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Could not get current dir due to %s\n", err)
			os.Exit(1)
		}

		templateDir := filepath.Join(currentDir, repo, "templates", namespace)

		_, err = os.Stat(templateDir)
		if err != nil {
			fmt.Printf("No %s dir in templates dir %s", namespace, err)
			os.Exit(1)
		}

		files, err := os.ReadDir(templateDir)
		if err != nil {
			fmt.Printf("Could not read dir due to %s\n", err)
			os.Exit(1)
		}

		if len(files) == 0 {
			fmt.Printf("Empty template namespace\n")
			os.Exit(1)
		}

		configFile, err := os.ReadFile(filepath.Join(currentDir, repo, fmt.Sprintf("config.%s", format)))
		if err != nil {
			fmt.Printf("Could not read config file due to %s\n", err)
			os.Exit(1)
		}

		var config map[string]any

		fileExtension := filepath.Ext(filepath.Join(currentDir, repo, fmt.Sprintf("config.%s", format)))

		switch fileExtension {
		case ".json":
			fmt.Printf("Found JSON config file\n")
			err = json.Unmarshal(configFile, &config)
			if err != nil {
				fmt.Printf("Could not unmarshal JSON config file %s\n", err)
				os.Exit(1)
			}
		case ".yaml", ".yml":
			fmt.Printf("Found YAML config file\n")
			err = yaml.Unmarshal(configFile, &config)
			if err != nil {
				fmt.Printf("Could not unmarshal YAML config file %s\n", err)
				os.Exit(1)
			}
		}

		for _, file := range files {

			templatePath := filepath.Join(templateDir, file.Name())

			content, err := os.ReadFile(templatePath)
			if err != nil {
				fmt.Printf("Could not read template file %s due to %s\n", file.Name(), err)
				continue
			}

			templ, err := template.New(file.Name()).Funcs(templates.ManiplacerFuncs).Parse(string(content))
			if err != nil {
				fmt.Printf("Could not parse %s due to %s\n", file.Name(), err)
				continue
			}

			outputDir := filepath.Join(currentDir, repo, "manifests", namespace, time.Now().Format("2006-01-02_15-04-05"))
			err = os.MkdirAll(outputDir, 0744)
			if err != nil {
				fmt.Printf("Could not create output dir %s due to %s\n", outputDir, err)
				os.Exit(1)
			}

			outputPath := filepath.Join(outputDir, file.Name())
			f, err := os.Create(outputPath)
			if err != nil {
				fmt.Printf("Could not create file %s due to %s\n", outputPath, err)
				continue
			}
			defer f.Close()

			err = templ.Execute(f, config)
			if err != nil {
				fmt.Printf("Could not execute template %s due to %s\n", file.Name(), err)
			} else {
				fmt.Printf("Generated %s\n", outputPath)
			}

		}

	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringP("format", "f", "json", "Config file format for generating manifest")
	generateCmd.Flags().StringP("namespace", "n", "default", "Namespace for template to be generated")
	generateCmd.Flags().StringP("repo", "r", "", "Repo name")
}
