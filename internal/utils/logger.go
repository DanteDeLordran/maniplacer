package utils

import (
	"context"
	"log/slog"
	"os"
)

var Log *slog.Logger

func init() {
	Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: getLogLevel(),
	}))
}

func getLogLevel() slog.Level {
	if os.Getenv("MANIPLACER_DEBUG") == "true" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

// Logger returns the global logger
func Logger() *slog.Logger {
	return Log
}

// LoggerFromContext returns a logger from context, or the global logger if not found
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return Log
}

type loggerKey struct{}

// ContextWithLogger returns a context with the logger attached
func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}
