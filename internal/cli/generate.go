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
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates the manifest given a config.json file and at least one template",
	Long: `Generates Kubernetes manifests from templates and a configuration file.

The 'generate' command reads a JSON configuration file (default: config.json) and one or more templates
from the 'templates' directory of your Maniplacer project. Each template is applied with the values from
the configuration file and generates a Kubernetes manifest.

Generated manifests are saved under 'manifests/<namespace>/<timestamp>/', where <namespace> is the
template namespace and <timestamp> ensures each run is stored in a unique folder.

Example usage:

  maniplacer generate
  maniplacer generate -n default
  maniplacer generate -f myconfig.json -n develop

Notes:

- The current directory must be a valid Maniplacer project (i.e., contain a '.maniplacer' file).
- Templates must exist under 'templates/<namespace>/'.
- Each run creates a timestamped output folder to avoid overwriting previous manifests.`,
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

		file, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Printf("Could not parse file flag, using default\n")
			file = "config.json"
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

		configFile, err := os.ReadFile(filepath.Join(currentDir, repo, file))
		if err != nil {
			fmt.Printf("Could not read config file due to %s\n", err)
			os.Exit(1)
		}

		var config map[string]any
		err = json.Unmarshal(configFile, &config)
		if err != nil {
			fmt.Printf("Could not unmarshal config file %s\n", err)
			os.Exit(1)
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
	generateCmd.Flags().StringP("file", "f", "config.json", "Config file for generating manifest")
	generateCmd.Flags().StringP("namespace", "n", "default", "Namespace for template to be generated")
	generateCmd.Flags().StringP("repo", "r", "", "Repo name")
}
