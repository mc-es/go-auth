package logger

import (
	"go-auth/pkg/logger/internal/core"
	"go-auth/pkg/logger/internal/provider"
)

// Shortcuts for logger types.
type (
	Logger         = provider.Logger
	Driver         = core.Driver
	Level          = core.Level
	Format         = core.Format
	TimeLayout     = core.TimeLayout
	FileRotation   = core.FileRotation
	ExtractCtxFunc = core.ExtractCtxFunc
)

// Shortcuts for logger configuration.
const (
	DriverZap     = core.Driver("zap")
	DriverLogrus  = core.Driver("logrus")
	DriverZerolog = core.Driver("zerolog")
	DriverNop     = core.Driver("nop")

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

	factory, err := provider.Get(cfg.Driver)
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
		Driver:      DriverZerolog,
		Level:       LevelInfo,
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
