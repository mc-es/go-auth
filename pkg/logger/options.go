package logger

import (
	"go-auth/pkg/logger/internal/core"
)

type Option func(*core.Config)

func WithDriver(driver Driver) Option {
	return func(cfg *core.Config) {
		cfg.Driver = driver
	}
}

func WithLevel(level Level) Option {
	return func(cfg *core.Config) {
		cfg.Level = level
	}
}

func WithFormat(format Format) Option {
	return func(cfg *core.Config) {
		cfg.Format = format
	}
}

func WithTimeLayout(timeLayout TimeLayout) Option {
	return func(cfg *core.Config) {
		cfg.TimeLayout = timeLayout
	}
}

func WithOutputPaths(outputPaths ...string) Option {
	return func(cfg *core.Config) {
		cfg.OutputPaths = outputPaths
	}
}

func WithDevelopment() Option {
	return func(cfg *core.Config) {
		cfg.Development = true
		cfg.Level = LevelDebug
		cfg.Format = FormatText
		cfg.TimeLayout = TimeLayoutTimeOnly
		cfg.OutputPaths = []string{"stdout"}
	}
}

func WithFileRotation(maxAge, maxSize, maxBackups int, localTime, compress bool) Option {
	return func(cfg *core.Config) {
		cfg.FileRotation = FileRotation{
			MaxAge:     maxAge,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			LocalTime:  localTime,
			Compress:   compress,
		}
	}
}

func WithContextExtractor(fn ExtractCtxFunc) Option {
	return func(cfg *core.Config) {
		cfg.Extractor = fn
	}
}
