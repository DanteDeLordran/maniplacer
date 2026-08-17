package cli

import (
	"fmt"
	"os"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		target, err := resolveTarget(cmd)
		if err != nil {
			return err
		}

		manifestsDir := target.ManifestsDir()

		if _, err := os.Stat(manifestsDir); err != nil {
			return fmt.Errorf("manifest directory does not exist: %w", err)
		}

		files, err := os.ReadDir(manifestsDir)
		if err != nil {
			return fmt.Errorf("could not read manifests directory: %w", err)
		}

		fmt.Printf("Manifests in %s namespace:\n", target.Namespace)
		for _, file := range files {
			fmt.Printf("- %s\n", file.Name())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	addRepoNamespaceFlags(listCmd, "Namespace for listing manifests")
}
