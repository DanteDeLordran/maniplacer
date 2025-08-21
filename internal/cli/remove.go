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

Examples:
  maniplacer remove Service -r myrepo
  maniplacer remove Service Deployment -n staging -r myrepo`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		if !utils.IsValidProject() {
			fmt.Printf("Current directory is not a valid Maniplacer project\n")
			os.Exit(1)
		}

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			fmt.Printf("Could not parse namespace flag due to %s, using default", err)
			namespace = "default"
		}

		repo, err := cmd.Flags().GetString("repo")
		if err != nil {
			fmt.Printf("Could not get repo flag due to %s\n", err)
			os.Exit(1)
		}

		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Could not get current directory due to %s", err)
			os.Exit(1)
		}

		templatesPath := filepath.Join(currentDir, repo, "templates", namespace)

		_, err = os.Stat(templatesPath)
		if err != nil {
			fmt.Printf("No namespace with name %s\n", namespace)
			os.Exit(1)
		}

		for _, comp := range args {
			templatePath := filepath.Join(templatesPath, fmt.Sprintf("%s.yaml", comp))

			file, err := os.Stat(templatePath)
			if err != nil {
				fmt.Printf("Component %s does not exist in templates dir with %s namespace, skipping...\n", comp, namespace)
			} else {
				fmt.Printf("Removing %s...\n", file.Name())

				err = os.Remove(templatePath)
				if err != nil {
					fmt.Printf("Could not remove file due to %s\n", err)
					continue
				}

				fmt.Printf("Successfully removed %s from %s namespace\n", file.Name(), namespace)
			}

		}

	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().StringP("namespace", "n", "default", "Namespace for removing templates")
	removeCmd.Flags().StringP("repo", "r", "", "Repo name")
}
