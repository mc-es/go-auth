package logger

import (
	"strings"

	"go-auth/pkg/logger/internal/core"
)

func validateConfig(cfg *core.Config) error {
	validators := []func(*core.Config) error{
		validateDriver,
		validateLevel,
		validateFormat,
		validateTimeLayout,
		validateOutputPaths,
		validateFileRotation,
	}

	for _, validate := range validators {
		if err := validate(cfg); err != nil {
			return err
		}
	}

	return nil
}

func validateDriver(cfg *core.Config) error {
	if cfg.Driver == Driver("") {
		return core.ErrMissingDriver
	}

	return nil
}

func validateLevel(cfg *core.Config) error {
	switch cfg.Level {
	case LevelDebug, LevelInfo, LevelWarn, LevelError, LevelPanic, LevelFatal:
		return nil
	default:
		return core.ErrInvalidLevel
	}
}

func validateFormat(cfg *core.Config) error {
	if cfg.Format != FormatJSON && cfg.Format != FormatText {
		return core.ErrInvalidFormat
	}

	return nil
}

func validateTimeLayout(cfg *core.Config) error {
	if cfg.TimeLayout == TimeLayout("") {
		return core.ErrInvalidTimeLayout
	}

	return nil
}

func validateOutputPaths(cfg *core.Config) error {
	if len(cfg.OutputPaths) == 0 {
		return core.ErrInvalidPaths
	}

	for _, path := range cfg.OutputPaths {
		if strings.TrimSpace(path) == "" {
			return core.ErrInvalidPaths
		}
	}

	return nil
}

func validateFileRotation(cfg *core.Config) error {
	if cfg.FileRotation.MaxAge <= 0 || cfg.FileRotation.MaxSize <= 0 || cfg.FileRotation.MaxBackups <= 0 {
		return core.ErrInvalidFileRotation
	}

	return nil
}
