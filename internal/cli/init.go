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

		if utils.IsValidProject() {
			fmt.Printf("Can not init a project since project already exists in current dir\n")
			os.Exit(1)
		}

		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		path, err := os.Getwd()

		if err != nil {
			fmt.Println("Could not get current dir, trying using $HOME")
			path, err = os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("could not use current dir nor $HOME due to: %w", err)
			}
		}

		if name == "" {

			confirm := utils.ConfirmMessage("No name given for project, do you want to use current dir?")
			if confirm {
				fmt.Println("Creating project on current dir...")
			} else {
				fmt.Printf("No project will be created :P\n")
				os.Exit(1)
			}

		}

		path = filepath.Join(path, name)

		if err := os.MkdirAll(path, 0744); err != nil {
			return fmt.Errorf("failed to create directory %w", err)
		}

		err = utils.CreateManiplacerProject(path)
		if err != nil {
			fmt.Printf("Error creating Maniplacer project file due to %s", err)
		}

		fmt.Println("Project initialized successfully.")

		confirm := utils.ConfirmMessage("Would you like to init a new repo inside your project? (You can create one later with maniplacer new <name>)")
		if confirm {
			name := getRepoName()

			dirs := []string{"", "templates", "manifests"}

			for _, dir := range dirs {
				if err := os.MkdirAll(filepath.Join(path, name, dir), 0744); err != nil {
					return fmt.Errorf("failed to create directory %w", err)
				}
			}

			if f, err := os.Create(filepath.Join(path, name, "config.json")); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			} else {
				defer f.Close()
			}

			fmt.Printf("Successfully created %s repo\n", name)

		} else {
			fmt.Println("Skipping repo init")
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
