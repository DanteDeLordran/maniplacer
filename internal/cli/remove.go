package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/dantedelordran/maniplacer/internal/templates"
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

		target, err := resolveTarget(cmd)
		if err != nil {
			return err
		}

		templatesPath := target.TemplatesDir()

		if _, err := os.Stat(templatesPath); err != nil {
			return fmt.Errorf("namespace '%s' does not exist", target.Namespace)
		}

		for _, comp := range args {
			if !slices.Contains(templates.AllowedComponents, comp) {
				return fmt.Errorf("unknown component %q", comp)
			}
			templatePath := filepath.Join(templatesPath, fmt.Sprintf("%s.yaml", comp))

			file, err := os.Stat(templatePath)
			if err != nil {
				logger.Debug("component does not exist, skipping", "component", comp, "namespace", target.Namespace)
				fmt.Printf("Component '%s' does not exist in templates dir with %s namespace, skipping...\n", comp, target.Namespace)
				continue
			}

			if err = os.Remove(templatePath); err != nil {
				logger.Warn("could not remove file", "file", file.Name(), "error", err)
				fmt.Printf("Could not remove file due to %s\n", err)
				continue
			}

			fmt.Printf("Successfully removed %s from %s namespace\n", file.Name(), target.Namespace)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	addRepoNamespaceFlags(removeCmd, "Namespace for removing templates")
}
