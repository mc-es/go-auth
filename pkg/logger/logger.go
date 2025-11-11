// Package logger wraps zap with sane defaults and global helpers
package logger

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger        // global logger instance
	globalSugar  *zap.SugaredLogger // cached sugared logger
	initOnce     sync.Once          // ensures Init runs only once
	initErr      error              // stores Init error for repeated calls
)

type config struct {
	// core
	level       zapcore.Level // debug, info, warn, error, fatal
	encoding    string        // json, console
	development bool          // enables dev-friendly settings

	// outputs
	outputPaths      []string // log output targets
	errorOutputPaths []string // error output targets

	// formatting
	timeEncoder   zapcore.TimeEncoder // custom time encoder
	initialFields map[string]any      // predefined structured fields

	// caller
	disableCaller bool // disable caller info
	callerSkip    int  // skip levels for wrapped loggers

	// stacktrace
	disableStacktrace bool          // disable stacktrace output
	stacktraceLevel   zapcore.Level // minimum level that triggers a stacktrace

	// sampling
	sampling           bool // enable log sampling
	samplingInitial    int  // log every message until this count
	samplingThereafter int  // log every Nth message afterward
}

type option func(*config) error

// New constructs a zap.Logger based on the provided options
func New(opts ...option) (*zap.Logger, error) {
	cfg, err := buildConfig(opts...)
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
	}

	zapCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(cfg.level),
		Development:       cfg.development,
		Encoding:          cfg.encoding,
		EncoderConfig:     buildEncoderConfig(cfg),
		OutputPaths:       cfg.outputPaths,
		ErrorOutputPaths:  cfg.errorOutputPaths,
		DisableCaller:     cfg.disableCaller,
		DisableStacktrace: cfg.disableStacktrace,
		InitialFields:     cfg.initialFields,
	}

	if cfg.sampling && cfg.samplingInitial > 0 && cfg.samplingThereafter > 0 {
		zapCfg.Sampling = &zap.SamplingConfig{
			Initial:    cfg.samplingInitial,
			Thereafter: cfg.samplingThereafter,
		}
	}

	var options []zap.Option
	if !cfg.disableStacktrace {
		options = append(options, zap.AddStacktrace(cfg.stacktraceLevel))
	}
	if cfg.callerSkip > 0 {
		options = append(options, zap.AddCallerSkip(cfg.callerSkip))
	}

	logger, err := zapCfg.Build(options...)
	if err != nil {
		return nil, fmt.Errorf("New(build): %w", err)
	}
	return logger, nil
}

// MustNew returns a configured logger or panics if setup fails
func MustNew(opts ...option) *zap.Logger {
	l, err := New(opts...)
	if err != nil {
		panic(fmt.Errorf("MustNew: %w", err))
	}
	return l
}

// Init initializes the package-level logger if it has not been created
func Init(opts ...option) error {
	initOnce.Do(func() {
		var err error
		globalLogger, err = New(opts...)
		if err != nil {
			initErr = fmt.Errorf("logger: Init: %w", err)
			return
		}
		globalSugar = globalLogger.Sugar()
	})
	return initErr
}

// L returns the global zap.Logger, panics if not initialized
func L() *zap.Logger {
	if logger := globalLogger; logger != nil {
		return logger
	}
	panic("logger.L(): not initialized, call logger.Init(...) at startup")
}

// S returns a global SugaredLogger derived from the base logger
func S() *zap.SugaredLogger {
	if sugar := globalSugar; sugar != nil {
		return sugar
	}
	panic("logger.S(): not initialized, call logger.Init(...) at startup")
}

// Sync flushes buffered log entries on the global logger
func Sync() error {
	if globalLogger == nil {
		return nil
	}
	if err := globalLogger.Sync(); err != nil && !isIgnorableSyncErr(err) {
		return err
	}
	return nil
}

// WithLevel overrides the minimum log level
func WithLevel(level string) option {
	return func(c *config) error {
		switch strings.ToLower(strings.TrimSpace(level)) {
		case "debug":
			c.level = zapcore.DebugLevel
		case "info":
			c.level = zapcore.InfoLevel
		case "warn":
			c.level = zapcore.WarnLevel
		case "error":
			c.level = zapcore.ErrorLevel
		case "fatal":
			c.level = zapcore.FatalLevel
		default:
			return fmt.Errorf("WithLevel: invalid level %q (allowed: debug, info, warn, error, fatal)", level)
		}
		return nil
	}
}

// WithEncoding sets the encoder to either json or console output
func WithEncoding(encoding string) option {
	return func(c *config) error {
		enc := strings.TrimSpace(strings.ToLower(encoding))
		switch enc {
		case "json", "console":
			c.encoding = enc
		default:
			return fmt.Errorf("WithEncoding: invalid encoding %q (allowed: json, console)", encoding)
		}
		return nil
	}
}

// WithDevelopmentMode enables zap dev mode with color console + short timestamps
func WithDevelopmentMode() option {
	return func(c *config) error {
		c.development = true
		c.encoding = "console"
		c.sampling = false
		if c.timeEncoder == nil {
			c.timeEncoder = zapcore.TimeEncoderOfLayout(time.TimeOnly)
		}
		return nil
	}
}

// WithOutputPaths replaces the main output destinations
func WithOutputPaths(paths ...string) option {
	return func(c *config) error {
		if len(paths) > 0 {
			c.outputPaths = append([]string{}, paths...)
		}
		return nil
	}
}

// WithErrorOutputPaths replaces the error output destinations
func WithErrorOutputPaths(paths ...string) option {
	return func(c *config) error {
		if len(paths) > 0 {
			c.errorOutputPaths = append([]string{}, paths...)
		}
		return nil
	}
}

// WithTimeEncoder customizes how timestamps are rendered
func WithTimeEncoder(enc zapcore.TimeEncoder) option {
	return func(c *config) error {
		if enc != nil {
			c.timeEncoder = enc
		}
		return nil
	}
}

// WithInitialFields attaches the given fields to every log entry
func WithInitialFields(fields map[string]any) option {
	return func(c *config) error {
		if len(fields) == 0 {
			return nil
		}
		maps.Copy(c.initialFields, fields)
		return nil
	}
}

// WithoutCaller disables caller metadata emission
func WithoutCaller() option {
	return func(c *config) error {
		c.disableCaller = true
		return nil
	}
}

// WithCallerSkip moves the reported call site up the stack
func WithCallerSkip(skip int) option {
	return func(c *config) error {
		if skip > 0 {
			c.callerSkip = skip
		}
		return nil
	}
}

// WithoutStacktrace disables stacktrace logging
func WithoutStacktrace() option {
	return func(c *config) error {
		c.disableStacktrace = true
		return nil
	}
}

// WithStacktraceLevel sets the minimum level that triggers a stacktrace
func WithStacktraceLevel(level string) option {
	return func(c *config) error {
		switch strings.ToLower(strings.TrimSpace(level)) {
		case "debug":
			c.stacktraceLevel = zapcore.DebugLevel
		case "info":
			c.stacktraceLevel = zapcore.InfoLevel
		case "warn":
			c.stacktraceLevel = zapcore.WarnLevel
		case "error":
			c.stacktraceLevel = zapcore.ErrorLevel
		case "fatal":
			c.stacktraceLevel = zapcore.FatalLevel
		default:
			return fmt.Errorf("WithStacktraceLevel: invalid level %q (allowed: debug, info, warn, error, fatal)", level)
		}
		return nil
	}
}

// WithSampling enables zap's sampling with custom thresholds
func WithSampling(initial, thereafter int) option {
	return func(c *config) error {
		if initial > 0 && thereafter > 0 {
			c.sampling = true
			c.samplingInitial = initial
			c.samplingThereafter = thereafter
		}
		return nil
	}
}

// WithoutSampling disables zap's log sampling entirely
func WithoutSampling() option {
	return func(c *config) error {
		c.sampling = false
		return nil
	}
}

func defaultConfig() config {
	return config{
		level:       zapcore.InfoLevel,
		encoding:    "json",
		development: false,

		outputPaths:      []string{"stdout"},
		errorOutputPaths: []string{"stderr"},

		timeEncoder:   nil,
		initialFields: map[string]any{},

		disableCaller: false,
		callerSkip:    0,

		disableStacktrace: false,
		stacktraceLevel:   zapcore.ErrorLevel,

		sampling:           true,
		samplingInitial:    100,
		samplingThereafter: 100,
	}
}

func buildConfig(opts ...option) (config, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&cfg); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

func buildEncoderConfig(cfg config) zapcore.EncoderConfig {
	encoder := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.DateTime),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.timeEncoder != nil {
		encoder.EncodeTime = cfg.timeEncoder
	}
	if cfg.development && cfg.encoding == "console" {
		encoder.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return encoder
}

func isIgnorableSyncErr(err error) bool {
	if err == nil {
		return false
	}

	var errno syscall.Errno
	if errors.As(err, &errno) && isIgnorableErrno(errno) {
		return true
	}

	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		if errno, ok := pathErr.Err.(syscall.Errno); ok && isIgnorableErrno(errno) {
			return true
		}
	}
	return false
}

func isIgnorableErrno(errno syscall.Errno) bool {
	switch errno {
	case syscall.ENOTTY, syscall.EINVAL, syscall.EPIPE, syscall.ENOSYS:
		return true
	default:
		return false
	}
}
