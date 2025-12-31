// Package zap provides zap logger adapter.
package zap

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go-auth/pkg/logger"
)

type adapter struct {
	logger *zap.SugaredLogger
	skip   int // skip levels for wrapped loggers
}

var _ logger.Logger = (*adapter)(nil)

// New creates a new zap logger instance.
func New(opts ...Option) (logger.Logger, error) {
	cfg, err := buildConfig(opts...)
	if err != nil {
		return nil, err
	}

	err = ensureDirectories(cfg.outputPaths)
	if err != nil {
		return nil, err
	}

	zapCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.Level(cfg.level)),
		Development:       cfg.development,
		Encoding:          string(cfg.encoding),
		EncoderConfig:     buildEncoderConfig(&cfg),
		OutputPaths:       cfg.outputPaths,
		ErrorOutputPaths:  cfg.errorOutputPaths,
		DisableCaller:     cfg.disableCaller,
		DisableStacktrace: cfg.disableStacktrace,
		InitialFields:     cfg.initialFields,
	}

	if cfg.sampling {
		zapCfg.Sampling = &zap.SamplingConfig{
			Initial:    cfg.samplingInitial,
			Thereafter: cfg.samplingThereafter,
		}
	}

	var options []zap.Option

	options = append(options, zap.AddCallerSkip(defaultCallerSkip))

	if !cfg.disableStacktrace {
		options = append(options, zap.AddStacktrace(zapcore.Level(cfg.stacktraceLevel)))
	}

	zapLogger, err := zapCfg.Build(options...)
	if err != nil {
		return nil, err
	}

	return &adapter{logger: zapLogger.Sugar(), skip: defaultCallerSkip}, nil
}

// Debug logs a debug message.
func (a *adapter) Debug(msg string, args ...any) { a.logger.Debugw(msg, args...) }

// Info logs an info message.
func (a *adapter) Info(msg string, args ...any) { a.logger.Infow(msg, args...) }

// Warn logs a warning message.
func (a *adapter) Warn(msg string, args ...any) { a.logger.Warnw(msg, args...) }

// Error logs an error message.
func (a *adapter) Error(msg string, args ...any) { a.logger.Errorw(msg, args...) }

// Fatal logs a fatal message and exits.
func (a *adapter) Fatal(msg string, args ...any) { a.logger.Fatalw(msg, args...) }

// With returns a logger with additional fields.
func (a *adapter) With(args ...any) logger.Logger {
	childLogger := a.logger.With(args...)

	childSkip := a.skip
	if a.skip > 1 {
		childLogger = childLogger.Desugar().WithOptions(zap.AddCallerSkip(-1)).Sugar()
		childSkip--
	}

	return &adapter{logger: childLogger, skip: childSkip}
}

// Sync flushes buffered log entries.
func (a *adapter) Sync() error {
	err := a.logger.Sync()
	if err != nil && !isIgnorableSyncErr(err) {
		return err
	}

	return nil
}

func defaultConfig() config {
	return config{
		level:       LevelDebug,
		encoding:    EncodingJson,
		development: false,

		outputPaths:      []string{"stdout"},
		errorOutputPaths: []string{"stderr"},

		timeEncoder:   nil,
		initialFields: map[string]any{},

		disableCaller: false,

		disableStacktrace: false,
		stacktraceLevel:   LevelError,

		sampling:           true,
		samplingInitial:    defaultSamplingInitial,
		samplingThereafter: defaultSamplingThereafter,
	}
}

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

func buildEncoderConfig(cfg *config) zapcore.EncoderConfig {
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

func ensureDirectories(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	for _, path := range paths {
		if path == "stdout" || path == "stderr" {
			continue
		}

		dir := filepath.Dir(path)

		if dir != "." && dir != "/" {
			if err := os.MkdirAll(dir, 0o750); err != nil {
				return err
			}
		}
	}

	return nil
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
		if errors.As(pathErr.Err, &errno) && isIgnorableErrno(errno) {
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
