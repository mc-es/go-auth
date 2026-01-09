package logger

import (
	"go-auth/pkg/logger/internal/core"
	"go-auth/pkg/logger/internal/driver"
)

// Shortcuts for logger types.
type (
	Logger         = driver.Logger
	Attr           = core.Attr
	Driver         = core.Driver
	Level          = core.Level
	Format         = core.Format
	TimeLayout     = core.TimeLayout
	FileRotation   = core.FileRotation
	ExtractCtxFunc = core.ExtractCtxFunc
)

// Shortcuts for attribute creation.
var (
	Str = core.String
	Int = core.Int
	Any = core.Any
)

// Shortcuts for logger configuration.
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
