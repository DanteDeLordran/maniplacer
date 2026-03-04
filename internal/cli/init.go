package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a project scaffolding",
	Long: `The init command bootstraps a new Maniplacer project by creating the required folder structure and configuration files.

It prepares the environment so you can immediately start adding and generating Kubernetes manifests. If a project name is provided with --name (or -n), it creates a new project folder with that name. Otherwise, it can initialize the project in the current working directory after confirmation.

During initialization, the following happens:
- A project root is created and registered as a valid Maniplacer project.
- The required directories are set up:
  * templates/   → for reusable Kubernetes component templates.
  * manifests/   → for generated manifests ready to be applied.
- A default config.json file is created in the project root.
- Optional: you can initialize a new repository inside the project, which sets up its own templates/ and manifests/ directories, along with a config.json file.

This makes it easy to start from a clean, organized structure without having to manually configure everything.

Example usage:
  maniplacer init
  maniplacer init --name my-app

After initialization, you can:
- Add components with 'maniplacer add'
- Generate manifests with 'maniplacer generate'
- Manage multiple repos inside the same project for different environments or services.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := utils.LoggerFromContext(cmd.Context())

		if utils.IsValidProject() {
			return fmt.Errorf("current directory is already a valid Maniplacer project")
		}

		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		path, err := os.Getwd()
		if err != nil {
			logger.Debug("could not get current directory, trying $HOME", "error", err)
			path, err = os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not use current dir nor $HOME: %w", err)
			}
		}

		if name == "" {
			confirm := utils.ConfirmMessage("No name given for project, do you want to use current dir?")
			if !confirm {
				fmt.Println("No project will be created")
				os.Exit(0)
			}
			logger.Info("creating project in current directory")
		} else {
			// Validate project name
			if err := utils.ValidateProjectName(name); err != nil {
				return fmt.Errorf("invalid project name: %w", err)
			}
			logger.Info("creating project", "name", name)
		}

		path = filepath.Join(path, name)

		if err := os.MkdirAll(path, utils.DirPermission); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		if err := utils.CreateManiplacerProject(path); err != nil {
			return fmt.Errorf("error creating Maniplacer project: %w", err)
		}

		logger.Info("project initialized successfully")

		confirm := utils.ConfirmMessage("Would you like to init a new repo inside your project? (You can create one later with maniplacer new <name>)")
		if confirm {
			repoName := getRepoName()

			// Validate repo name
			if err := utils.ValidateRepoName(repoName); err != nil {
				return fmt.Errorf("invalid repository name: %w", err)
			}

			dirs := []string{"", "templates", "manifests"}

			for _, dir := range dirs {
				if err := os.MkdirAll(filepath.Join(path, repoName, dir), utils.DirPermission); err != nil {
					return fmt.Errorf("failed to create directory: %w", err)
				}
			}

			if f, err := os.Create(filepath.Join(path, repoName, "config.json")); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			} else {
				defer f.Close()
			}

			logger.Info("repository created successfully", "name", repoName)
		} else {
			logger.Info("skipping repository initialization")
		}

		fmt.Printf("Run 'cd %s'\n", path)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func getRepoName() string {
	fmt.Printf("What would be the name of your repo?: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Could not read input due to: %s\n", err)
		os.Exit(1)
	}
	input = strings.TrimSpace(strings.ToLower(input))
	return input
}
