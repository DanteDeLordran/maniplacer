package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Creates a new repo in project",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		name := args[0]

		if !utils.IsValidProject() {
			fmt.Printf("Not a valid Maniplacer project\n")
			os.Exit(1)
		}

		currentPath, err := os.Getwd()
		if err != nil {
			fmt.Printf("Could not get current dir due to %s\n", err)
			os.Exit(1)
		}

		repoPath := filepath.Join(currentPath, name)

		if err = os.MkdirAll(repoPath, 0744); err != nil {
			fmt.Printf("Could not create %s repo %s\n", name, err)
			os.Exit(1)
		}

		dirs := []string{"", "manifests", "templates"}

		for _, dir := range dirs {
			if err = os.MkdirAll(filepath.Join(repoPath, dir), 0744); err != nil {
				fmt.Printf("Could not create %s repo %s\n", name, err)
				os.Exit(1)
			}
		}

		if f, err := os.Create(filepath.Join(repoPath, "config.json")); err != nil {
			fmt.Printf("failed to create config file: %s", err)
			os.Exit(1)
		} else {
			defer f.Close()
		}

		fmt.Printf("Successfuly created repo %s inside your project\n", name)

	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
