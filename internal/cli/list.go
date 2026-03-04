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

		manifestsDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get current directory: %w", err)
		}

		manifestsDir = filepath.Join(manifestsDir, repo, "manifests", namespace)

		_, err = os.Stat(manifestsDir)
		if err != nil {
			return fmt.Errorf("manifest directory does not exist: %w", err)
		}

		files, err := os.ReadDir(manifestsDir)
		if err != nil {
			return fmt.Errorf("could not read manifests directory: %w", err)
		}

		logger.Info("listing manifests", "namespace", namespace, "count", len(files))
		fmt.Printf("Manifests in %s namespace:\n", namespace)
		for _, file := range files {
			fmt.Printf("- %s\n", file.Name())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("namespace", "n", utils.DefaultNamespace, "Namespace for listing manifests")
	listCmd.Flags().StringP("repo", "r", "", "Repo name")
}
