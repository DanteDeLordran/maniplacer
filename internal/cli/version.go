/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"fmt"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

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
	Long: `Displays the current version of Maniplacer along with the project logo.

The version is defined internally at build time, allowing you to quickly
verify which release of Maniplacer you are running.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\nVersion: %s\n", logo, getVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func getVersion() string {
	return utils.Version
}
