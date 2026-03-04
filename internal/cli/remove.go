package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Removes a component from the templates dir given a namespace, defaults to default namespace",
	Long: `Removes one or more components from the templates directory in the given namespace.
If no namespace is specified, the "default" namespace is used.

This command deletes the corresponding YAML files under "templates/<namespace>"
inside the specified repository.

Supported components include:
- Service
- Deployment
- HttpRoute
- Secret
- ConfigMap
- HPA
- HCPolicy

Examples:
  maniplacer remove service -r myrepo
  maniplacer remove service deployment -n staging -r myrepo`,
	Args: cobra.MinimumNArgs(1),
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

		templatesPath := filepath.Join(currentDir, repo, "templates", namespace)

		_, err = os.Stat(templatesPath)
		if err != nil {
			return fmt.Errorf("namespace '%s' does not exist", namespace)
		}

		for _, comp := range args {
			templatePath := filepath.Join(templatesPath, fmt.Sprintf("%s.yaml", comp))

			file, err := os.Stat(templatePath)
			if err != nil {
				logger.Warn("component does not exist, skipping", "component", comp, "namespace", namespace)
				fmt.Printf("Component '%s' does not exist in templates dir with %s namespace, skipping...\n", comp, namespace)
				continue
			}

			logger.Info("removing component", "component", comp, "namespace", namespace)
			fmt.Printf("Removing %s...\n", file.Name())

			if err = os.Remove(templatePath); err != nil {
				logger.Warn("could not remove file", "file", file.Name(), "error", err)
				fmt.Printf("Could not remove file due to %s\n", err)
				continue
			}

			logger.Info("component removed", "component", comp, "namespace", namespace)
			fmt.Printf("Successfully removed %s from %s namespace\n", file.Name(), namespace)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().StringP("namespace", "n", utils.DefaultNamespace, "Namespace for removing templates")
	removeCmd.Flags().StringP("repo", "r", "", "Repo name")
}
