package logger

import (
	"strings"

	"go-auth/pkg/logger/internal/core"
)

type validator struct {
	cfg *core.Config
}

func validateConfig(cfg *core.Config) error {
	validator := &validator{cfg: cfg}
	validators := []func() error{
		validator.validateDriver,
		validator.validateLevel,
		validator.validateFormat,
		validator.validateTimeLayout,
		validator.validateOutputPaths,
		validator.validateFileRotation,
	}

	for _, validate := range validators {
		if err := validate(); err != nil {
			return err
		}
	}

	return nil
}

func (v *validator) validateDriver() error {
	if v.cfg.Driver == Driver("") {
		return core.ErrMissingDriver
	}

	return nil
}

func (v *validator) validateLevel() error {
	switch v.cfg.Level {
	case LevelDebug, LevelInfo, LevelWarn, LevelError, LevelPanic, LevelFatal:
		return nil
	default:
		return core.ErrInvalidLevel
	}
}

func (v *validator) validateFormat() error {
	if v.cfg.Format != FormatJSON && v.cfg.Format != FormatText {
		return core.ErrInvalidFormat
	}

	return nil
}

func (v *validator) validateTimeLayout() error {
	if v.cfg.TimeLayout == TimeLayout("") {
		return core.ErrInvalidTimeLayout
	}

	return nil
}

func (v *validator) validateOutputPaths() error {
	if len(v.cfg.OutputPaths) == 0 {
		return core.ErrInvalidPaths
	}

	for _, path := range v.cfg.OutputPaths {
		if strings.TrimSpace(path) == "" {
			return core.ErrInvalidPaths
		}
	}

	return nil
}

func (v *validator) validateFileRotation() error {
	if v.cfg.FileRotation.MaxAge <= 0 || v.cfg.FileRotation.MaxSize <= 0 || v.cfg.FileRotation.MaxBackups <= 0 {
		return core.ErrInvalidFileRotation
	}

	return nil
}
