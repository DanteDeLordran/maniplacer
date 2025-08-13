package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a project scaffolding",
	Long:  `Initializes the scaffolding for a new Maniplacer project`,
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

		fmt.Println("Project initialized successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("name", "n", "", "Name of the new project")
}
