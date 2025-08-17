package cli

import (
	"fmt"
	"os"
	"path/filepath"

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

		if name != "" {
			path = filepath.Join(path, name)
			fmt.Println("Creating project on: ", path)
		} else {
			fmt.Println("Creating project on current dir")
		}

		dirs := []string{"", "templates", "manifests"}
		for _, d := range dirs {
			if err := os.MkdirAll(filepath.Join(path, d), 0744); err != nil {
				return fmt.Errorf("failed to create directory %q: %w", d, err)
			}
		}

		if f, err := os.Create(filepath.Join(path, "config.json")); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		} else {
			defer f.Close()
		}

		err = utils.CreateManiplacerProject(path)

		if err != nil {
			fmt.Printf("Error creating Maniplacer project file due to %s", err)
		}

		fmt.Println("Project initialized successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("name", "n", "", "Name of the new project")
}
