package cli

import (
	"context"
	"fmt"
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
	// A command that fails at runtime is not a usage mistake, so don't dump the
	// help text after it; Execute reports the error once instead.
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger with context
		cmd.SetContext(utils.ContextWithLogger(cmd.Context(), utils.Logger()))
	},
}

func Execute() {
	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	// Enable shell completion
	rootCmd.CompletionOptions.DisableDefaultCmd = false
}
