package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deployedCmd = &cobra.Command{
	Use:   "deployed",
	Short: "Lists the pods that where deployed using Maniplacer",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deployed called")
	},
}

func init() {
	rootCmd.AddCommand(deployedCmd)
}
