// Package logger builds the application's structured logger. It's the only
// place log output is configured; everything else logs via slog.Default()
// (populated by Init) so no logger needs to be threaded through call sites.
package logger

import (
	"log/slog"
	"os"
)

// Init builds a JSON slog.Logger writing to stdout, installs it as the
// package-level default via slog.SetDefault, and returns it for callers
// (e.g. main.go) that want to log directly rather than through the default.
//
// Level is controlled by the LOG_LEVEL env var ("debug", "warn", "error";
// anything else, including unset, defaults to "info").
func Init() *slog.Logger {
	level := parseLevel(os.Getenv("LOG_LEVEL"))
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger
}

func parseLevel(raw string) slog.Level {
	switch raw {
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
