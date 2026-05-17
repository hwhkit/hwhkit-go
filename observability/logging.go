// Package observability wires logging, metrics, and tracing for hwhkit-go services.
package observability

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/hwhkit/hwhkit-go/config"
)

func InitLogging(cfg config.LogConfig) *slog.Logger {
	return InitLoggingTo(cfg, os.Stderr)
}

func InitLoggingTo(cfg config.LogConfig, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(cfg.Level)}
	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "text":
		handler = slog.NewTextHandler(w, opts)
	default:
		handler = slog.NewJSONHandler(w, opts)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
