/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "1.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Returns the current version of Maniplacer",
	Long:  `Returns the current version of Maniplacer`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", getVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func getVersion() string {
	return version
}
