package logger

import (
	"log/slog"
	"os"
)

const (
	dev   = "dev"
	prod  = "prod"
	local = "local"
)

func InitLogger(level string) *slog.Logger {
	var newLog *slog.Logger

	switch level {
	case local:
		newLog = slog.New(slog.NewTextHandler(
			os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug},
		))
	case dev:
		newLog = slog.New(slog.NewJSONHandler(
			os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug},
		))
	case prod:
		newLog = slog.New(slog.NewJSONHandler(
			os.Stdout, &slog.HandlerOptions{Level: slog.LevelError},
		))
	}

	return newLog
}
