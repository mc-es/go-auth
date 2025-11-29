// Package logger wraps zap with sane defaults and global helpers.
package logger

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New constructs a Logger based on the provided options.
func New(opts ...Option) (Logger, error) {
	cfg, err := buildConfig(opts...)
	if err != nil {
		return nil, fmt.Errorf("New: %w", err)
	}

	zapCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.Level(cfg.level)),
		Development:       cfg.development,
		Encoding:          string(cfg.encoding),
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
		options = append(options, zap.AddStacktrace(zapcore.Level(cfg.stacktraceLevel)))
	}

	if cfg.callerSkip > 0 {
		options = append(options, zap.AddCallerSkip(cfg.callerSkip))
	}

	logger, err := zapCfg.Build(options...)
	if err != nil {
		return nil, fmt.Errorf("New(build): %w", err)
	}

	return &zapLogger{Logger: logger}, nil
}

// NewSugar constructs a SugarLogger based on the provided options.
func NewSugar(opts ...Option) (SugarLogger, error) {
	logger, err := New(opts...)
	if err != nil {
		return nil, fmt.Errorf("NewSugar: %w", err)
	}

	zapL, ok := logger.(*zapLogger)
	if !ok {
		return nil, fmt.Errorf("NewSugar: unexpected logger type")
	}

	return &zapSugarLogger{SugaredLogger: zapL.zap().Sugar()}, nil
}

// Init initializes the package-level logger if it has not been created.
func Init(opts ...Option) error {
	initOnce.Do(func() {
		var err error

		globalLogger, err = New(opts...)
		if err != nil {
			errInit = fmt.Errorf("logger: Init: %w", err)

			return
		}

		zapL, ok := globalLogger.(*zapLogger)
		if !ok {
			errInit = fmt.Errorf("logger: Init: unexpected logger type")

			return
		}

		globalSugar = &zapSugarLogger{SugaredLogger: zapL.zap().Sugar()}
	})

	return errInit
}

// L returns the global Logger, panics if not initialized.
func L() Logger {
	if logger := globalLogger; logger != nil {
		return logger
	}

	panic("logger.L(): not initialized, call logger.Init(...) at startup")
}

// S returns the global SugarLogger, panics if not initialized.
func S() SugarLogger {
	if sugar := globalSugar; sugar != nil {
		return sugar
	}

	panic("logger.S(): not initialized, call logger.Init(...) at startup")
}

// Sync flushes buffered log entries on the global logger.
func Sync() error {
	if globalLogger == nil {
		return nil
	}

	if err := globalLogger.Sync(); err != nil && !isIgnorableSyncErr(err) {
		return err
	}

	return nil
}

// defaultConfig seeds a config with standard production-ready defaults.
func defaultConfig() config {
	return config{
		level:       LevelInfo,
		encoding:    EncodingJson,
		development: false,

		outputPaths:      []string{"stdout"},
		errorOutputPaths: []string{"stderr"},

		timeEncoder:   nil,
		initialFields: map[string]any{},

		disableCaller: false,
		callerSkip:    0,

		disableStacktrace: false,
		stacktraceLevel:   LevelError,

		sampling:           true,
		samplingInitial:    defaultSamplingInitial,
		samplingThereafter: defaultSamplingThereafter,
	}
}

// buildConfig applies a sequence of Option mutators over the defaultConfig.
func buildConfig(opts ...Option) (config, error) {
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

// buildEncoderConfig maps the config values to zap's encoder settings.
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

	if cfg.development && cfg.encoding == EncodingConsole {
		encoder.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	return encoder
}

// isIgnorableSyncErr filters out known benign errors from zap.Sync().
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
		if errors.As(pathErr.Err, &errno) && isIgnorableErrno(errno) {
			return true
		}
	}

	return false
}

// isIgnorableErrno reports whether the provided errno is harmless for sync operations.
func isIgnorableErrno(errno syscall.Errno) bool {
	switch errno {
	case syscall.ENOTTY, syscall.EINVAL, syscall.EPIPE, syscall.ENOSYS:
		return true
	default:
		return false
	}
}

// zap returns the underlying *zap.Logger from zapLogger.
func (z *zapLogger) zap() *zap.Logger {
	return z.Logger
}
