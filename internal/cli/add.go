/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
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

type component string

const (
	deployment component = "deployment"
	service    component = "service"
	configmap  component = "configmap"
)

var allowedComponents = []string{string(deployment), string(service), string(configmap)}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a new K8s component manifest file placeholder to your Maniplacer project for you to customize",
	Long:  `The add command lets you create a new component placeholder and add it to your project`,
	Args:  cobra.MinimumNArgs(1),
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

		for _, comp := range args {
			if slices.Contains(allowedComponents, comp) {
				fmt.Printf("Creating %s in templates directory in %s namespace\n", comp, namespace)

				d := templates.DeploymentTemplate{}.Deployment()

				current, err := os.Getwd()
				if err != nil {
					fmt.Printf("Could not get current dir due to %s\n", err)
					os.Exit(1)
				}

				outputDir := filepath.Join(current, "templates", fmt.Sprintf("%s.yaml", comp))

				if err := os.WriteFile(outputDir, d, 0644); err != nil {
					fmt.Printf("failed to write file: %s\n", err)
					os.Exit(1)
				}

				fmt.Printf("%s.yaml successfully generated!\n", comp)

			} else {
				fmt.Printf("No component with name %s, skipping...\n", comp)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringP("namespace", "n", "default", "Namespace for your component template")
}
