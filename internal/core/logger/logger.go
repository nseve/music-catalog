package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	*slog.Logger
	file *os.File
}

func New(config Config) (*slog.Logger, func() error, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, nil, fmt.Errorf("unmarshal log level: %w", err)
	}

	if err := os.MkdirAll(config.Folder, 0755); err != nil {
		return nil, nil, fmt.Errorf("mkdir log folder: %w", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000000")
	logFilePath := filepath.Join(config.Folder, fmt.Sprintf("%s.log", timestamp))

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}

	w := io.MultiWriter(os.Stdout, logFile)

	h := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(h), logFile.Close, nil
}
