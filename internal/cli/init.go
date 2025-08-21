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
	Long: `The init command initializes the scaffolding for a new Maniplacer project.

It sets up the required project directory structure and configuration files 
so you can start managing Kubernetes manifests immediately.

Specifically, it will:
- Create a project root directory (if --name is provided, it will use that name, otherwise the current directory).
- Generate the required subdirectories: 
  * templates/   → for storing reusable Kubernetes manifest templates.
  * manifests/   → for placing your customized manifests ready to be applied.
- Create a default config.json file in the project root.
- Register the project as a valid Maniplacer project.

Example usage:
  maniplacer init
  maniplacer init --name my-app

After initialization, you can run 'maniplacer add' to add Kubernetes components
and start customizing them for your project.`,
	Args: cobra.MaximumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {

		if utils.IsValidProject() {
			fmt.Printf("Can not init a project since project already exists in current dir\n")
			os.Exit(1)
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Println("Could not get name flag: ", err)
			return err
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

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("name", "n", "", "Name of the new project")
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
