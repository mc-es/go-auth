package logrus

import (
	"io"
	"maps"
	"time"
)

// WithLevel sets the log level. Default: -1 (debug).
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

// WithFormatter sets the log formatter. Default: json.
//
//	"text" | "json".
func WithFormatter(formatter Formatter) Option {
	return func(cfg *config) {
		switch formatter {
		case FormatterText, FormatterJSON:
			cfg.formatter = formatter
		default:
			cfg.formatter = FormatterJSON
		}
	}
}

// WithOutput sets the log output writer. Default: os.Stdout.
func WithOutput(output io.Writer) Option {
	return func(cfg *config) {
		if output != nil {
			cfg.output = output
		}
	}
}

// WithDevelopmentMode enables development-friendly settings.
//
//	development: true, formatter: text, timeLayout: time.TimeOnly.
func WithDevelopmentMode() Option {
	return func(cfg *config) {
		cfg.development = true
		cfg.formatter = FormatterText
		cfg.timeLayout = time.TimeOnly
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
