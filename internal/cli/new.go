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
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.LoggerFromContext(cmd.Context())
		name := args[0]

		// Validate repo name
		if err := utils.ValidateRepoName(name); err != nil {
			return fmt.Errorf("invalid repository name: %w", err)
		}

		// Check for path traversal
		if err := utils.ValidateSafePath(name); err != nil {
			return err
		}

		if !utils.IsValidProject() {
			return fmt.Errorf("current directory is not a valid Maniplacer project")
		}

		currentPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get current directory: %w", err)
		}

		repoPath := filepath.Join(currentPath, name)

		// Check if repo already exists
		if _, err := os.Stat(repoPath); err == nil {
			return fmt.Errorf("repository '%s' already exists", name)
		}

		if err = os.MkdirAll(repoPath, utils.DirPermission); err != nil {
			return fmt.Errorf("could not create %s repo: %w", name, err)
		}

		dirs := []string{"", "manifests", "templates"}

		for _, dir := range dirs {
			if err = os.MkdirAll(filepath.Join(repoPath, dir), utils.DirPermission); err != nil {
				return fmt.Errorf("could not create %s repo: %w", name, err)
			}
		}

		configPath := filepath.Join(repoPath, "config.json")
		if f, err := os.Create(configPath); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		} else {
			defer f.Close()
		}

		logger.Info("repository created successfully", "name", name, "path", repoPath)
		fmt.Printf("Successfully created repo '%s' inside your project\n", name)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
