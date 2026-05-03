// Package observability bündelt Logging, Metriken und (perspektivisch)
// Tracing-Helfer für das Backend.
package observability

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type ctxKey struct{}

var loggerKey ctxKey

// NewLogger erstellt einen slog.Logger gemäß Format ("json"|"text") und Level
// ("debug"|"info"|"warn"|"error"). Schreibt nach os.Stdout.
func NewLogger(format, level string) *slog.Logger {
	return newLoggerWithWriter(os.Stdout, format, level)
}

func newLoggerWithWriter(w io.Writer, format, level string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}

	var h slog.Handler
	if format == "text" {
		h = slog.NewTextHandler(w, opts)
	} else {
		h = slog.NewJSONHandler(w, opts)
	}
	return slog.New(h)
}

func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithLogger hängt einen Logger an den Context.
func WithLogger(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

// FromContext liefert den Logger aus dem Context oder den Default-Logger.
func FromContext(ctx context.Context) *slog.Logger {
	if log, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return log
	}
	return slog.Default()
}
