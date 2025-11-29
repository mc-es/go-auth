package logger

import (
	"maps"
	"time"

	"go.uber.org/zap/zapcore"
)

// WithLevel overrides the minimum log level.
func WithLevel(level level) Option {
	return func(cfg *config) error {
		cfg.level = level

		return nil
	}
}

// WithEncoding sets the encoder to either json or console output.
func WithEncoding(encoding encoding) Option {
	return func(cfg *config) error {
		cfg.encoding = encoding

		return nil
	}
}

// WithDevelopmentMode enables zap dev mode with color console + short timestamps.
func WithDevelopmentMode() Option {
	return func(cfg *config) error {
		cfg.development = true
		cfg.encoding = EncodingConsole
		cfg.sampling = false
		cfg.timeEncoder = zapcore.TimeEncoderOfLayout(time.TimeOnly)

		return nil
	}
}

// WithOutputPaths replaces the main output destinations.
func WithOutputPaths(paths ...string) Option {
	return func(cfg *config) error {
		if len(paths) > 0 {
			cfg.outputPaths = make([]string, len(paths))
			copy(cfg.outputPaths, paths)
		}

		return nil
	}
}

// WithErrorOutputPaths replaces the error output destinations.
func WithErrorOutputPaths(paths ...string) Option {
	return func(cfg *config) error {
		if len(paths) > 0 {
			cfg.errorOutputPaths = make([]string, len(paths))
			copy(cfg.errorOutputPaths, paths)
		}

		return nil
	}
}

// WithTimeEncoder customizes how timestamps are rendered.
func WithTimeEncoder(enc zapcore.TimeEncoder) Option {
	return func(cfg *config) error {
		if enc != nil {
			cfg.timeEncoder = enc
		}

		return nil
	}
}

// WithInitialFields attaches the given fields to every log entry.
func WithInitialFields(fields map[string]any) Option {
	return func(cfg *config) error {
		if len(fields) == 0 {
			return nil
		}

		maps.Copy(cfg.initialFields, fields)

		return nil
	}
}

// WithoutCaller disables caller metadata emission.
func WithoutCaller() Option {
	return func(cfg *config) error {
		cfg.disableCaller = true

		return nil
	}
}

// WithCallerSkip moves the reported call site up the stack.
func WithCallerSkip(skip int) Option {
	return func(cfg *config) error {
		if skip > 0 {
			cfg.callerSkip = skip
		}

		return nil
	}
}

// WithoutStacktrace disables stacktrace logging.
func WithoutStacktrace() Option {
	return func(cfg *config) error {
		cfg.disableStacktrace = true

		return nil
	}
}

// WithStacktraceLevel sets the minimum level that triggers a stacktrace.
func WithStacktraceLevel(level level) Option {
	return func(cfg *config) error {
		cfg.stacktraceLevel = level

		return nil
	}
}

// WithSampling enables zap's sampling with custom thresholds.
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

// WithoutSampling disables zap's log sampling entirely.
func WithoutSampling() Option {
	return func(cfg *config) error {
		cfg.sampling = false

		return nil
	}
}
