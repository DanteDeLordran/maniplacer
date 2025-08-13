package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "maniplacer",
	Short: "Maniplacer CLI for generating K8s manifests",
	Long: `Maniplacer is a CLI tool for generating K8s manifests based on a config file and templates, similar to Helm but simplier
	
It generates the manifest in your local project in order for you to apply or store as you like
	`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

}
