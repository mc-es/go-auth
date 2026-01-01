package zap

import (
	"maps"
	"time"
)

// WithLevel sets the log level. Default: -1.
//
//	"debug": -1 | "info": 0 | "warn": 1 | "error": 2 | "fatal": 3.
func WithLevel(level Level) Option {
	return func(cfg *config) {
		switch level {
		case LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal:
			cfg.level = level
		default:
			cfg.level = LevelDebug
		}
	}
}

// WithEncoding sets the log encoding format. Default: json.
//
//	"json" | "console".
func WithEncoding(encoding Encoding) Option {
	return func(cfg *config) {
		switch encoding {
		case EncodingJson, EncodingConsole:
			cfg.encoding = encoding
		default:
			cfg.encoding = EncodingJson
		}
	}
}

// WithDevelopmentMode enables development-friendly settings.
//
//	development: true, encoding: console, sampling: false, timeLayout: time.TimeOnly.
func WithDevelopmentMode() Option {
	return func(cfg *config) {
		cfg.development = true
		cfg.encoding = EncodingConsole
		cfg.sampling = false
		cfg.timeLayout = time.TimeOnly
	}
}

// WithOutputPaths sets the log output paths. Default: stdout.
func WithOutputPaths(paths ...string) Option {
	return func(cfg *config) {
		if len(paths) > 0 {
			cfg.outputPaths = make([]string, len(paths))
			copy(cfg.outputPaths, paths)
		}
	}
}

// WithTimeLayout sets a custom time layout. Default: time.DateTime.
//
//	time.DateOnly | time.TimeOnly | time.RFC3339 | time.RFC822 | time.RFC1123.
func WithTimeLayout(layout string) Option {
	return func(cfg *config) {
		switch layout {
		case time.DateTime, time.DateOnly, time.TimeOnly, time.RFC3339, time.RFC822, time.RFC1123:
			cfg.timeLayout = layout
		default:
			cfg.timeLayout = time.DateTime
		}
	}
}

// WithInitialFields sets fields added to all log entries.
//
// Example:
//
//	WithInitialFields(map[string]any{
//		"key": "value",
//	})
func WithInitialFields(fields map[string]any) Option {
	return func(cfg *config) {
		maps.Copy(cfg.initialFields, fields)
	}
}

// WithoutCaller disables caller information in logs.
func WithoutCaller() Option {
	return func(cfg *config) {
		cfg.disableCaller = true
	}
}

// WithoutStacktrace disables stacktrace output.
func WithoutStacktrace() Option {
	return func(cfg *config) {
		cfg.disableStacktrace = true
	}
}

// WithoutSampling disables log sampling.
func WithoutSampling() Option {
	return func(cfg *config) {
		cfg.sampling = false
	}
}

// WithSampling enables log sampling with given parameters. (positive values).
func WithSampling(initial, thereafter int) Option {
	return func(cfg *config) {
		if initial > 0 && thereafter > 0 {
			cfg.sampling = true
			cfg.samplingInitial = initial
			cfg.samplingThereafter = thereafter
		}
	}
}
