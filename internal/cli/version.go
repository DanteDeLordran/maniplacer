/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "1.0.0"

const logo = `
                       _       _                     
 _ __ ___   __ _ _ __ (_)_ __ | | __ _  ___ ___ _ __ 
| '_ ' _ \ / _' | '_ \| | '_ \| |/ _' |/ __/ _ \ '__|
| | | | | | (_| | | | | | |_) | | (_| | (_|  __/ |   
|_| |_| |_|\__,_|_| |_|_| .__/|_|\__,_|\___\___|_|   
                        |_|                                                         
`

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Returns the current version of Maniplacer",
	Long:  `Returns the current version of Maniplacer`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\nVersion: %s\n", logo, getVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func getVersion() string {
	return version
}
