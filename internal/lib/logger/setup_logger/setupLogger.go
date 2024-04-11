package setupLogger

import (
	"log/slog"
	"os"
)

const (
	logLVLInfo  = "info"
	logLVLDebug = "debug"
	logLVLWarn  = "warning"
	logLVLError = "error"
)

func New(logLVL string) *slog.Logger {
	var log *slog.Logger

	switch logLVL {
	case logLVLInfo:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case logLVLDebug:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case logLVLWarn:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	case logLVLError:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	}

	return log
}
