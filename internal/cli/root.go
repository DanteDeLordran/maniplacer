package cli

import (
	"context"
	"os"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "maniplacer",
	Short: "Maniplacer CLI for generating K8s manifests",
	Long: `Maniplacer is a CLI tool for generating K8s manifests based on a config file and templates, similar to Helm but simpler.

It generates the manifest in your local project in order for you to apply or store as you like.
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger with context
		ctx := context.WithValue(cmd.Context(), loggerKey{}, utils.Logger())
		cmd.SetContext(ctx)
	},
}

type loggerKey struct{}

func Execute() {
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		utils.Logger().Error("command execution failed", "error", err)
		os.Exit(1)
	}
}

func init() {
	// Enable shell completion
	rootCmd.CompletionOptions.DisableDefaultCmd = false
}
