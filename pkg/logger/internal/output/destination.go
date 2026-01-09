package output

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"

	"go-auth/pkg/logger/internal/core"
)

type Destination struct {
	Console io.Writer
	File    io.Writer
	closers []io.Closer
}

func New(cfg *core.Config) (*Destination, error) {
	dest := &Destination{
		closers: make([]io.Closer, 0, len(cfg.OutputPaths)),
	}

	consolePaths, filePaths, err := partitionPaths(cfg.OutputPaths)
	if err != nil {
		return nil, err
	}

	if len(consolePaths) > 0 {
		writer, err := dest.openWriters(consolePaths, core.FileRotation{})
		if err != nil {
			_ = dest.Close()

			return nil, err
		}

		dest.Console = writer
	}

	if len(filePaths) > 0 {
		writer, err := dest.openWriters(filePaths, cfg.FileRotation)
		if err != nil {
			_ = dest.Close()

			return nil, err
		}

		dest.File = writer
	}

	return dest, nil
}

func (d *Destination) Close() error {
	var firstErr error
	for _, c := range d.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (d *Destination) openWriters(paths []string, rot core.FileRotation) (io.Writer, error) {
	var writers []io.Writer

	for _, path := range paths {
		switch path {
		case "stdout":
			writers = append(writers, os.Stdout)
		case "stderr":
			writers = append(writers, os.Stderr)
		default:
			if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
				return nil, fmt.Errorf("failed to create log directory: %w", err)
			}

			jack := &lumberjack.Logger{
				Filename:   path,
				MaxAge:     rot.MaxAge,
				MaxSize:    rot.MaxSize,
				MaxBackups: rot.MaxBackups,
				LocalTime:  rot.LocalTime,
				Compress:   rot.Compress,
			}

			writers = append(writers, jack)
			d.closers = append(d.closers, jack)
		}
	}

	return io.MultiWriter(writers...), nil
}

func partitionPaths(paths []string) (console, files []string, err error) {
	seen := make(map[string]struct{})

	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			return nil, nil, fmt.Errorf("log output path cannot be empty")
		}

		if path == "stdout" || path == "stderr" {
			if _, exists := seen[path]; !exists {
				seen[path] = struct{}{}
				console = append(console, path)
			}

			continue
		}

		absPath, err := resolveFile(path)
		if err != nil {
			return nil, nil, err
		}

		if _, exists := seen[absPath]; !exists {
			seen[absPath] = struct{}{}
			files = append(files, absPath)
		}
	}

	return console, files, nil
}

func resolveFile(path string) (string, error) {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %q: %w", path, err)
	}

	if info, err := os.Stat(abs); err == nil && info.IsDir() {
		return "", fmt.Errorf("log path is a directory: %s", abs)
	}

	return abs, nil
}
