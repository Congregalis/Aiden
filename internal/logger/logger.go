package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/congregalis/aiden/internal/config"
)

func New(cfg config.LogConfig) *slog.Logger {
	options := &slog.HandlerOptions{
		Level:     parseLevel(cfg.Level),
		AddSource: cfg.AddSource,
	}

	handler := slog.NewJSONHandler(os.Stdout, options)
	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
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
