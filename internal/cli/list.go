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
	Long: `The list command displays all generated manifests stored under a specific namespace in your Maniplacer project. 

It scans the 'manifests/<namespace>/' directory of the selected repository and prints out the manifest files available. By default, it looks in the 'default' namespace, but you can override this with the --namespace (or -n) flag. You must also specify the target repository with the --repo (or -r) flag.

This is useful for quickly checking which manifests are currently available for a given environment or namespace without manually browsing directories.

Examples:
  maniplacer list
  maniplacer list -n staging -r myrepo
  maniplacer list --namespace production --repo backend-service

Notes:
- The current directory must be a valid Maniplacer project (contain a '.maniplacer' file).
- The target namespace must already have generated manifests to be listed.`,
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
