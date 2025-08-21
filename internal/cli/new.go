package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Creates a new repo in project",
	Long: `The new command creates a fresh repository inside your existing Maniplacer project. 

A repository represents an isolated environment or service within your project, complete with its own configuration and directories for templates and manifests. This allows you to manage multiple environments (e.g., staging, production) or different services within the same project.

When you run 'maniplacer new <name>', it will:
- Create a new folder named after the repository.
- Initialize the required subdirectories:
  * templates/   → for reusable Kubernetes resource templates.
  * manifests/   → for generated manifests.
- Create a default config.json file inside the repository root.

Example usage:
  maniplacer new frontend
  maniplacer new backend

This will set up independent repos 'frontend' and 'backend' inside your project, each with their own templates, manifests, and config.json.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		name := args[0]

		if !utils.IsValidProject() {
			fmt.Printf("Not a valid Maniplacer project\n")
			os.Exit(1)
		}

		currentPath, err := os.Getwd()
		if err != nil {
			fmt.Printf("Could not get current dir due to %s\n", err)
			os.Exit(1)
		}

		repoPath := filepath.Join(currentPath, name)

		if err = os.MkdirAll(repoPath, 0744); err != nil {
			fmt.Printf("Could not create %s repo %s\n", name, err)
			os.Exit(1)
		}

		dirs := []string{"", "manifests", "templates"}

		for _, dir := range dirs {
			if err = os.MkdirAll(filepath.Join(repoPath, dir), 0744); err != nil {
				fmt.Printf("Could not create %s repo %s\n", name, err)
				os.Exit(1)
			}
		}

		if f, err := os.Create(filepath.Join(repoPath, "config.json")); err != nil {
			fmt.Printf("failed to create config file: %s", err)
			os.Exit(1)
		} else {
			defer f.Close()
		}

		fmt.Printf("Successfuly created repo %s inside your project\n", name)

	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
