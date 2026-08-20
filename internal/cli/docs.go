package cli

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/dantedelordran/maniplacer/internal/utils"
	"github.com/spf13/cobra"
)

//go:embed docs.html
var docsPage []byte

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Serve Maniplacer documentation locally",
	Long: `The docs command serves Maniplacer's documentation from a loopback-only HTTP server.

By default, the server uses port 8000. Override it with --port (or -p), then
open the reported /docs/ URL in a browser.

Example:
  maniplacer docs --port 9000`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		portValue, err := cmd.Flags().GetString("port")
		if err != nil {
			return fmt.Errorf("could not parse port flag: %w", err)
		}

		port, err := parseDocsPort(portValue)
		if err != nil {
			return err
		}

		return runDocsServer(cmd.Context(), port, cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.Flags().StringP("port", "p", utils.DefaultPort, "Port for serving the docs page (1-65535)")
}

func parseDocsPort(value string) (int, error) {
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid port %q: must be an integer from 1 to 65535", value)
	}
	return port, nil
}

func docsHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("GET /docs/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(docsPage)
	})
	return mux
}

func runDocsServer(ctx context.Context, port int, output io.Writer) error {
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return fmt.Errorf("failed to bind documentation server: %w", err)
	}

	server := &http.Server{
		Handler:           docsHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/docs/", port)
	utils.LoggerFromContext(ctx).Info("starting documentation server", "url", url)
	fmt.Fprintf(output, "Documentation server available at: %s\n", url)

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("documentation server failed: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("failed to shut down documentation server: %w", err)
		}
		if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("documentation server failed during shutdown: %w", err)
		}
		return nil
	}
}
