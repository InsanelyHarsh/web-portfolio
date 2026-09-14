package logger

import (
	"log/slog"
	"os"
)

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
