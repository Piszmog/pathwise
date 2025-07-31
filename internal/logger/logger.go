package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

func New(logLevel, logOutput string) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	var writer io.Writer = os.Stdout
	if logOutput != "" && logOutput != "stdout" {
		writer = &lumberjack.Logger{
			Filename:   logOutput,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}
	}

	return slog.New(
		slog.NewTextHandler(
			writer,
			&slog.HandlerOptions{Level: level},
		),
	)
}
