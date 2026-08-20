package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

// target identifies the repository and namespace a command operates on,
// rooted at the current Maniplacer project directory.
type target struct {
	Root      string
	Repo      string
	Namespace string
}

// RepoPath returns the repository root inside the project
func (t target) RepoPath() string {
	return filepath.Join(t.Root, t.Repo)
}

// TemplatesDir returns the directory holding the namespace's templates
func (t target) TemplatesDir() string {
	return filepath.Join(t.RepoPath(), "templates", t.Namespace)
}

// ManifestsDir returns the directory holding the namespace's generated manifests
func (t target) ManifestsDir() string {
	return filepath.Join(t.RepoPath(), "manifests", t.Namespace)
}

// resolveTarget validates that the current directory is a Maniplacer project and
// resolves the --repo and --namespace flags shared by every repo-scoped command.
func resolveTarget(cmd *cobra.Command) (target, error) {
	logger := utils.LoggerFromContext(cmd.Context())

	if !utils.IsValidProject() {
		return target{}, fmt.Errorf("current directory is not a valid Maniplacer project")
	}

	namespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		logger.Debug("could not parse namespace flag, using default", "error", err)
		namespace = utils.DefaultNamespace
	}

	if err := utils.ValidateNamespace(namespace); err != nil {
		return target{}, fmt.Errorf("invalid namespace: %w", err)
	}

	repo, err := cmd.Flags().GetString("repo")
	if err != nil {
		return target{}, fmt.Errorf("could not get repo flag: %w", err)
	}

	if repo == "" {
		return target{}, fmt.Errorf("repository name is required (use --repo flag)")
	}

	if err := utils.ValidateRepoName(repo); err != nil {
		return target{}, fmt.Errorf("invalid repository name: %w", err)
	}
	if err := utils.ValidateSafePath(repo); err != nil {
		return target{}, err
	}

	root, err := os.Getwd()
	if err != nil {
		return target{}, fmt.Errorf("could not get current directory: %w", err)
	}

	return target{Root: root, Repo: repo, Namespace: namespace}, nil
}

// addRepoNamespaceFlags registers the --namespace/--repo flag pair that every
// repo-scoped command shares.
func addRepoNamespaceFlags(cmd *cobra.Command, namespaceUsage string) {
	cmd.Flags().StringP("namespace", "n", utils.DefaultNamespace, namespaceUsage)
	cmd.Flags().StringP("repo", "r", "", "Repo name")
}
