package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
	Version:       utils.Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger with context
		cmd.SetContext(utils.ContextWithLogger(cmd.Context(), utils.Logger()))
	},
}

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	// Enable shell completion
	rootCmd.CompletionOptions.DisableDefaultCmd = false
}
