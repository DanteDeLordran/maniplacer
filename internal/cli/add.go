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
	Long: `The add command scaffolds new Kubernetes component manifests and places them in your Maniplacer project’s templates directory. 

It allows you to generate one or multiple component YAML files at once, so you don’t have to manually write boilerplate configurations every time you add a resource. 
This saves time and ensures your manifests follow a consistent structure.

By default, generated manifests are placed under the "default" namespace, but you can override this with the --namespace (or -n) flag. 
You can also specify the target repository directory with the --repo (or -r) flag to control where the files are created.

Available components:
- Service       (Expose your application as a network-accessible service)
- Deployment    (Define workloads with containers, replicas, and rollout strategy)
- HttpRoute     (Configure HTTP routing rules for ingress traffic)
- Secret        (Securely store sensitive values like tokens, passwords, and certificates)
- ConfigMap     (Provide configuration values and environment variables as key-value pairs)

If a file already exists, you will be prompted to confirm before overwriting it, preventing accidental data loss.

Example usage:
  maniplacer add deployment service -n staging -r myrepo

This command generates deployment.yaml and service.yaml in the 
templates/staging directory of the "myrepo" project.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		if !utils.IsValidProject() {
			fmt.Printf("Current directory is not a valid Maniplacer project\n")
			os.Exit(1)
		}

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			fmt.Printf("Could not get namespace flag due to %s, using default namespace\n", err)
			namespace = "default"
		}

		repo, err := cmd.Flags().GetString("repo")
		if err != nil {
			fmt.Printf("Could not get repo flag due to %s\n", err)
			os.Exit(1)
		}

		current, err := os.Getwd()
		if err != nil {
			fmt.Printf("Could not get current dir due to %s\n", err)
			os.Exit(1)
		}

		repoPath := filepath.Join(current, repo)

		for _, comp := range args {
			if slices.Contains(templates.AllowedComponents, comp) {
				fmt.Printf("Creating %s in templates directory in %s namespace...\n", comp, namespace)

				t := templates.TemplateRegistry[comp]

				err = os.MkdirAll(filepath.Join(repoPath, "templates", namespace), 0744)
				if err != nil {
					fmt.Printf("Could not create templates namespace dir due to %s\n", err)
					os.Exit(1)
				}

				outputDir := filepath.Join(repoPath, "templates", namespace, fmt.Sprintf("%s.yaml", comp))

				if _, err := os.Stat(outputDir); err == nil {
					// File exists
					confirmed := utils.ConfirmMessage(fmt.Sprintf("%s already exists, do you want to replace it?", filepath.Base(outputDir)))
					if !confirmed {
						fmt.Printf("Skipping %s...\n", filepath.Base(outputDir))
						continue
					}
				} else if !os.IsNotExist(err) {
					// Some unexpected error (e.g. permission issue)
					fmt.Printf("Error checking file %s: %s\n", outputDir, err)
					os.Exit(1)
				}

				// Either the file does not exist, or user confirmed overwrite
				if err := os.WriteFile(outputDir, t, 0644); err != nil {
					fmt.Printf("Failed to write file: %s\n", err)
					os.Exit(1)
				}

				fmt.Printf("%s.yaml successfully generated in %s namespace!\n", comp, namespace)

			} else {
				fmt.Printf("No component with name %s, skipping...\n", comp)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringP("namespace", "n", "default", "Namespace for your component template")
	addCmd.Flags().StringP("repo", "r", "", "Repo name")
}
