package setupLogger

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const (
	logLVLInfo  = "info"
	logLVLDebug = "debug"
	logLVLWarn  = "warning"
	logLVLError = "error"
)

func New(logLVL string) (*slog.Logger, *os.File) {
	todayDate := time.Now().Format(time.DateOnly)
	logPath, err := filepath.Abs(fmt.Sprintf("/var/go-apps/corp-portal/prod/portal/logs/%s.txt", todayDate))
	if err != nil {
		panic(err)
	}
	var logFile *os.File

	// check if log file exists
	_, err = os.Stat(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			logFile, err = os.Create(logPath)
			if err != nil {
				panic(err)
			}
			err = logFile.Chmod(777)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_RDWR, 777) // os.O_APPEND - ставит текущую позицию каретки (чтения/записи в конец файла) os.O_RDWD - открывает для чтения и записи
		if err != nil {
			panic(err)
		}
	}

	var log *slog.Logger

	switch logLVL {
	case logLVLInfo:
		log = slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelInfo}))
	case logLVLDebug:
		log = slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case logLVLWarn:
		log = slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelWarn}))
	case logLVLError:
		log = slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelError}))
	}

	return log, logFile
}
