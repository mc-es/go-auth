package logger

import (
	"slices"

	"go-auth/pkg/logger/internal/core"
	"go-auth/pkg/logger/internal/driver"
)

type Logger = driver.Logger

// A is a shortcut for creating attributes.
var A = core.NewAttr

// Public constants for logger configuration.
const (
	DriverZap    = core.Driver("zap")
	DriverLogrus = core.Driver("logrus")

	LevelDebug = core.LevelDebug
	LevelInfo  = core.LevelInfo
	LevelWarn  = core.LevelWarn
	LevelError = core.LevelError
	LevelPanic = core.LevelPanic
	LevelFatal = core.LevelFatal

	FormatJSON = core.FormatJSON
	FormatText = core.FormatText

	TimeLayoutDateTime = core.TimeLayoutDateTime
	TimeLayoutDateOnly = core.TimeLayoutDateOnly
	TimeLayoutTimeOnly = core.TimeLayoutTimeOnly
	TimeLayoutRFC3339  = core.TimeLayoutRFC3339
	TimeLayoutRFC822   = core.TimeLayoutRFC822
	TimeLayoutRFC1123  = core.TimeLayoutRFC1123
)

func New(opts ...Option) (Logger, error) {
	cfg := defaultConfig()

	for _, opt := range opts {
		opt(&cfg)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	factory, err := driver.Get(cfg.Driver)
	if err != nil {
		return nil, err
	}

	return factory(&cfg)
}

func defaultConfig() core.Config {
	const (
		maxAge     = 7
		maxSize    = 100
		maxBackups = 3
		localTime  = true
		compress   = true
	)

	return core.Config{
		Driver:      DriverZap,
		Level:       LevelDebug,
		Format:      FormatJSON,
		TimeLayout:  TimeLayoutDateTime,
		OutputPaths: []string{"stdout"},
		Development: false,
		FileRotation: core.FileRotation{
			MaxAge:     maxAge,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			LocalTime:  localTime,
			Compress:   compress,
		},
	}
}

func validateConfig(cfg *core.Config) error {
	if cfg.Driver == "" {
		return core.ErrMissingDriver
	}

	switch cfg.Level {
	case core.LevelDebug, core.LevelInfo, core.LevelWarn, core.LevelError, core.LevelPanic, core.LevelFatal:
		break
	default:
		return core.ErrInvalidLevel
	}

	if cfg.Format != core.FormatJSON && cfg.Format != core.FormatText {
		return core.ErrInvalidFormat
	}

	if len(cfg.OutputPaths) == 0 || slices.Contains(cfg.OutputPaths, "") {
		return core.ErrInvalidPaths
	}

	return nil
}
