package storage

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"

	errs "github.com/nseve/music-catalog/internal/core/errors"
)

type Local struct {
	root *os.Root
	log  *slog.Logger
}

func NewLocal(dir string, log *slog.Logger) (*Local, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", dir, err)
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open root: %w", err)
	}

	return &Local{
		root: root,
		log:  log,
	}, nil
}

func (s *Local) Close() error {
	return s.root.Close()
}

func (s *Local) Save(key string, src io.Reader) error {
	f, err := s.root.OpenFile(key, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	if _, err := io.Copy(f, src); err != nil {
		_ = f.Close()
		_ = s.root.Remove(key)
		return fmt.Errorf("write file: %w", err)
	}

	if err := f.Close(); err != nil {
		_ = s.root.Remove(key)
		return fmt.Errorf("close file: %w", err)
	}

	return nil
}

func (s *Local) Open(key string) (*os.File, error) {
	f, err := s.root.Open(key)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("file %s: %w", key, errs.ErrNotFound)
	}

	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return f, nil
}

func (s *Local) Discard(key string) {
	if err := s.root.Remove(key); err != nil && !errors.Is(err, fs.ErrNotExist) {
		s.log.Warn(
			"failed to discard file",
			slog.String("key", key),
			slog.Any("error", err),
		)
	}
}
