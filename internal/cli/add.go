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

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a new K8s component manifest file placeholder to your Maniplacer project for you to customize",
	Long: `The add command scaffolds new Kubernetes component manifests and places them in your Maniplacer project's templates directory.

It allows you to generate one or multiple component YAML files at once, so you don't have to manually write boilerplate configurations every time you add a resource.
This saves time and ensures your manifests follow a consistent structure.

By default, generated manifests are placed under the "default" namespace, but you can override this with the --namespace (or -n) flag.
You can also specify the target repository directory with the --repo (or -r) flag to control where the files are created.

Available components:
- Service       (Expose your application as a network-accessible service)
- Deployment    (Define workloads with containers, replicas, and rollout strategy)
- HttpRoute     (Configure HTTP routing rules for ingress traffic)
- Secret        (Securely store sensitive values like tokens, passwords, and certificates)
- ConfigMap     (Provide configuration values and environment variables as key-value pairs)
- HPA           (Horizontal Pod Autoscaler for automatic scaling)
- HCPolicy      (Health Check Policy configuration)

If a file already exists, you will be prompted to confirm before overwriting it, preventing accidental data loss.

Example usage:
  maniplacer add deployment service -n staging -r myrepo

This command generates deployment.yaml and service.yaml in the
templates/staging directory of the "myrepo" project.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.LoggerFromContext(cmd.Context())

		if !utils.IsValidProject() {
			return fmt.Errorf("current directory is not a valid Maniplacer project")
		}

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			logger.Debug("could not get namespace flag, using default", "error", err)
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

		current, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get current directory: %w", err)
		}

		repoPath := filepath.Join(current, repo)

		// Validate repo path exists
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			return fmt.Errorf("repository '%s' does not exist", repo)
		}

		for _, comp := range args {
			if slices.Contains(templates.AllowedComponents, comp) {
				logger.Info("creating component template", "component", comp, "namespace", namespace)

				t := templates.TemplateRegistry[comp]

				templateDir := filepath.Join(repoPath, "templates", namespace)
				if err := os.MkdirAll(templateDir, utils.DirPermission); err != nil {
					return fmt.Errorf("could not create templates namespace directory: %w", err)
				}

				outputPath := filepath.Join(templateDir, fmt.Sprintf("%s.yaml", comp))

				if _, err := os.Stat(outputPath); err == nil {
					// File exists
					confirmed := utils.ConfirmMessage(fmt.Sprintf("%s already exists, do you want to replace it?", filepath.Base(outputPath)))
					if !confirmed {
						logger.Info("skipping component", "component", comp)
						fmt.Printf("Skipping %s...\n", filepath.Base(outputPath))
						continue
					}
				} else if !os.IsNotExist(err) {
					// Some unexpected error (e.g. permission issue)
					return fmt.Errorf("error checking file %s: %w", outputPath, err)
				}

				// Either the file does not exist, or user confirmed overwrite
				if err := os.WriteFile(outputPath, t, utils.FilePermission); err != nil {
					return fmt.Errorf("failed to write file: %w", err)
				}

				logger.Info("component template created", "component", comp, "namespace", namespace)
				fmt.Printf("%s.yaml successfully generated in %s namespace!\n", comp, namespace)

			} else {
				logger.Warn("unknown component, skipping", "component", comp)
				fmt.Printf("No component with name '%s', skipping...\n", comp)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringP("namespace", "n", utils.DefaultNamespace, "Namespace for your component template")
	addCmd.Flags().StringP("repo", "r", "", "Repo name")
}
