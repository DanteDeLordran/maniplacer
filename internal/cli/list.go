package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists every manifest from a given namespace",
	Long: `The list command displays all manifest files in the specified namespace.

By default, it lists manifests in the 'default' namespace.
Manifests are expected to be located inside the 'manifests/<namespace>' directory
of a valid Maniplacer project.

Examples:
  maniplacer list
  maniplacer list --namespace production`,
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

		repo, err := cmd.Flags().GetString("repo")
		if err != nil {
			fmt.Printf("Could not get repo flag due to %s\n", err)
			os.Exit(1)
		}

		manifestsDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Could not get current dir due to %s\n", err)
			os.Exit(1)
		}

		manifestsDir = filepath.Join(manifestsDir, repo, "manifests", namespace)

		_, err = os.Stat(manifestsDir)
		if err != nil {
			fmt.Printf("Manifest dir does not exists %s\n", err)
			os.Exit(1)
		}

		files, err := os.ReadDir(manifestsDir)
		if err != nil {
			fmt.Printf("Could not read manifests dir due to %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Manifests in %s namespace:\n", namespace)
		for _, file := range files {
			fmt.Printf("- %s\n", file.Name())
		}

	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("namespace", "n", "default", "Namespace for listing manifests")
	listCmd.Flags().StringP("repo", "r", "", "Repo name")
}
