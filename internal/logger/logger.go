package logger

import (
	"log/slog"
	"os"
)

const (
	prodEnv  = "prod"
	devEnv   = "dev"
	localEnv = "local"
)

func MustSetup(env string) *slog.Logger {
	switch env {
	case prodEnv:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}),
		)
	case devEnv:
		return slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case localEnv:
		return slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	default:
		return slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)

	}
}
