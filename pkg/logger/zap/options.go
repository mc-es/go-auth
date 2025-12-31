package zap

import (
	"maps"
	"time"

	"go.uber.org/zap/zapcore"
)

// WithLevel sets the log level. (-1: debug, 0: info, 1: warn, 2: error, 3: fatal).
func WithLevel(level Level) Option {
	return func(cfg *config) error {
		cfg.level = level

		return nil
	}
}

// WithEncoding sets the log encoding format. (json, console).
func WithEncoding(encoding Encoding) Option {
	return func(cfg *config) error {
		cfg.encoding = encoding

		return nil
	}
}

// WithDevelopmentMode enables development-friendly settings.
// (development: true, encoding: console, sampling: false, timeEncoder: time.TimeOnly).
func WithDevelopmentMode() Option {
	return func(cfg *config) error {
		cfg.development = true
		cfg.encoding = EncodingConsole
		cfg.sampling = false
		cfg.timeEncoder = zapcore.TimeEncoderOfLayout(time.TimeOnly)

		return nil
	}
}

// WithOutputPaths sets the log output paths.
func WithOutputPaths(paths ...string) Option {
	return func(cfg *config) error {
		if len(paths) > 0 {
			cfg.outputPaths = make([]string, len(paths))
			copy(cfg.outputPaths, paths)
		}

		return nil
	}
}

// WithErrorOutputPaths sets the error output paths.
func WithErrorOutputPaths(paths ...string) Option {
	return func(cfg *config) error {
		if len(paths) > 0 {
			cfg.errorOutputPaths = make([]string, len(paths))
			copy(cfg.errorOutputPaths, paths)
		}

		return nil
	}
}

// WithTimeEncoder sets a custom time encoder.
func WithTimeEncoder(enc zapcore.TimeEncoder) Option {
	return func(cfg *config) error {
		if enc != nil {
			cfg.timeEncoder = enc
		}

		return nil
	}
}

// WithInitialFields sets fields added to all log entries.
func WithInitialFields(fields map[string]any) Option {
	return func(cfg *config) error {
		if len(fields) == 0 {
			return nil
		}

		maps.Copy(cfg.initialFields, fields)

		return nil
	}
}

// WithoutCaller disables caller information in logs.
func WithoutCaller() Option {
	return func(cfg *config) error {
		cfg.disableCaller = true

		return nil
	}
}

// WithoutStacktrace disables stacktrace output.
func WithoutStacktrace() Option {
	return func(cfg *config) error {
		cfg.disableStacktrace = true

		return nil
	}
}

// WithStacktraceLevel sets the minimum level for stacktraces. (-1: debug, 0: info, 1: warn, 2: error, 3: fatal).
func WithStacktraceLevel(level Level) Option {
	return func(cfg *config) error {
		cfg.stacktraceLevel = level

		return nil
	}
}

// WithSampling enables log sampling with given parameters.
func WithSampling(initial, thereafter int) Option {
	return func(cfg *config) error {
		if initial > 0 && thereafter > 0 {
			cfg.sampling = true
			cfg.samplingInitial = initial
			cfg.samplingThereafter = thereafter
		}

		return nil
	}
}

// WithoutSampling disables log sampling.
func WithoutSampling() Option {
	return func(cfg *config) error {
		cfg.sampling = false

		return nil
	}
}
