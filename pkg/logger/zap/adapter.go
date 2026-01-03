// Package zap is the adapter for the zap logger.
package zap

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go-auth/pkg/logger"
)

type adapter struct {
	l *zap.SugaredLogger
}

// Register registers the zap adapter.
func Register() {
	logger.Register(logger.DriverZap, newZap)
}

func newZap(config *logger.Config) logger.Logger {
	consolePaths, filePaths := splitPaths(config.OutputPath)

	var cores []zapcore.Core

	if len(consolePaths) > 0 {
		consoleEncoder := buildEncoder(config, true)
		consoleWriter := buildWriteSyncer(consolePaths)
		cores = append(cores, zapcore.NewCore(consoleEncoder, consoleWriter, zapcore.Level(config.Level)))
	}

	if len(filePaths) > 0 {
		fileEncoder := buildEncoder(config, false)
		fileWriter := buildWriteSyncer(filePaths)
		cores = append(cores, zapcore.NewCore(fileEncoder, fileWriter, zapcore.Level(config.Level)))
	}

	core := zapcore.NewTee(cores...)

	opts := buildOptions(config)
	l := zap.New(core, opts...)

	return &adapter{l: l.Sugar()}
}

// Debug logs a debug message.
func (a *adapter) Debug(message string, args ...any) { a.l.Debugw(message, args...) }

// Info logs an info message.
func (a *adapter) Info(message string, args ...any) { a.l.Infow(message, args...) }

// Warn logs a warning message.
func (a *adapter) Warn(message string, args ...any) { a.l.Warnw(message, args...) }

// Error logs an error message.
func (a *adapter) Error(message string, args ...any) { a.l.Errorw(message, args...) }

// Panic logs a panic message.
func (a *adapter) Panic(message string, args ...any) { a.l.Panicw(message, args...) }

// Fatal logs a fatal message.
func (a *adapter) Fatal(message string, args ...any) { a.l.Fatalw(message, args...) }

// With returns a logger with additional fields.
func (a *adapter) With(args ...any) logger.Logger {
	return &adapter{l: a.l.With(args...)}
}

// Sync flushes buffered log entries.
func (a *adapter) Sync() error {
	return a.l.Sync()
}

func splitPaths(paths []string) (console, files []string) {
	if len(paths) == 0 {
		return []string{"stdout"}, nil
	}

	for _, p := range paths {
		if p == "stdout" || p == "stderr" {
			console = append(console, p)
		} else {
			files = append(files, p)
		}
	}

	return console, files
}

func buildEncoder(cfg *logger.Config, allowColor bool) zapcore.Encoder {
	enCfg := zap.NewProductionEncoderConfig()

	if cfg.DevMode {
		enCfg = zap.NewDevelopmentEncoderConfig()
	}

	enCfg.TimeKey = "time"
	enCfg.LevelKey = "level"
	enCfg.NameKey = "logger"
	enCfg.CallerKey = "caller"
	enCfg.MessageKey = "message"
	enCfg.StacktraceKey = "stacktrace"
	enCfg.EncodeTime = zapcore.TimeEncoderOfLayout(string(cfg.TimeLayout))

	if cfg.Formatter == logger.FormatterText && allowColor && cfg.DevMode {
		enCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		enCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	if cfg.Formatter == logger.FormatterJSON {
		return zapcore.NewJSONEncoder(enCfg)
	}

	return zapcore.NewConsoleEncoder(enCfg)
}

func buildOptions(cfg *logger.Config) []zap.Option {
	opts := []zap.Option{
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCallerSkip(1),
	}

	if !cfg.DisableCaller {
		opts = append(opts, zap.WithCaller(true))
	}

	return opts
}

func buildWriteSyncer(paths []string) zapcore.WriteSyncer {
	if len(paths) == 0 {
		return zapcore.AddSync(os.Stdout)
	}

	writers := make([]zapcore.WriteSyncer, 0, len(paths))

	for _, path := range paths {
		if path == "stdout" {
			writers = append(writers, zapcore.AddSync(os.Stdout))

			continue
		}

		if path == "stderr" {
			writers = append(writers, zapcore.AddSync(os.Stderr))

			continue
		}

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			panic(err)
		}

		file, err := os.OpenFile(filepath.Clean(path), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			panic(err)
		}

		writers = append(writers, zapcore.AddSync(file))
	}

	if len(writers) == 1 {
		return writers[0]
	}

	return zapcore.NewMultiWriteSyncer(writers...)
}
