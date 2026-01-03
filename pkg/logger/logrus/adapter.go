// Package logrus provides logrus logger adapter.
package logrus

import (
	"bytes"
	"path/filepath"
	"runtime"
	"strconv"

	"go-auth/pkg/logger"

	"github.com/sirupsen/logrus"
)

type adapter struct {
	logger            *logrus.Logger
	entry             *logrus.Entry
	disableStacktrace bool
}

var _ logger.Logger = (*adapter)(nil)

// fullLevelTextFormatter extends TextFormatter to show full level names.
type fullLevelTextFormatter struct {
	*logrus.TextFormatter
}

func (f *fullLevelTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Call parent Format
	data, err := f.TextFormatter.Format(entry)
	if err != nil {
		return nil, err
	}

	// Replace abbreviated level names with full names
	levelMap := map[string]string{
		"DEBU": "DEBUG",
		"INFO": "INFO",
		"WARN": "WARN",
		"ERRO": "ERROR",
		"FATA": "FATAL",
		"PANI": "PANIC",
	}

	for abbrev, full := range levelMap {
		data = bytes.ReplaceAll(data, []byte(abbrev), []byte(full))
	}

	return data, nil
}

// New creates a new logrus logger instance.
//
// Example:
//
//	logger, err := logrus.New(
//		logrus.WithLevel(logrus.LevelInfo),
//		logrus.WithFormatter(logrus.FormatterText),
//		... // other options
//	)
func New(opts ...Option) (logger.Logger, error) {
	cfg := buildConfig(opts...)

	l := logrus.New()
	l.SetFormatter(buildFormatter(&cfg))
	l.SetLevel(convertLevel(cfg.level))
	l.SetOutput(cfg.output)
	l.SetReportCaller(!cfg.disableCaller)

	entry := l.WithFields(logrus.Fields(cfg.initialFields))

	return &adapter{
		logger:            l,
		entry:             entry,
		disableStacktrace: cfg.disableStacktrace,
	}, nil
}

// Debug logs a debug message.
func (a *adapter) Debug(msg string, args ...any) {
	a.entry.WithFields(convertArgs(args...)).Debug(msg)
}

// Info logs an info message.
func (a *adapter) Info(msg string, args ...any) {
	a.entry.WithFields(convertArgs(args...)).Info(msg)
}

// Warn logs a warning message.
func (a *adapter) Warn(msg string, args ...any) {
	a.entry.WithFields(convertArgs(args...)).Warn(msg)
}

// Error logs an error message.
func (a *adapter) Error(msg string, args ...any) {
	a.entry.WithFields(convertArgs(args...)).Error(msg)
}

// Fatal logs a fatal message and exits.
func (a *adapter) Fatal(msg string, args ...any) {
	a.entry.WithFields(convertArgs(args...)).Fatal(msg)
}

// With returns a logger with additional fields.
func (a *adapter) With(args ...any) logger.Logger {
	fields := convertArgs(args...)
	newEntry := a.entry.WithFields(fields)

	return &adapter{
		logger:            a.logger,
		entry:             newEntry,
		disableStacktrace: a.disableStacktrace,
	}
}

// Sync flushes buffered log entries.
// Logrus doesn't buffer by default, so this is a no-op.
func (a *adapter) Sync() error {
	return nil
}

// convertArgs converts variadic key-value pairs to logrus.Fields.
// It expects args to be in the format: key1, value1, key2, value2, ...
func convertArgs(args ...any) logrus.Fields {
	fields := make(logrus.Fields)
	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		fields[key] = args[i+1]
	}

	return fields
}

// convertLevel converts our Level type to logrus.Level.
func convertLevel(level Level) logrus.Level {
	switch level {
	case LevelDebug:
		return logrus.DebugLevel
	case LevelInfo:
		return logrus.InfoLevel
	case LevelWarn:
		return logrus.WarnLevel
	case LevelError:
		return logrus.ErrorLevel
	case LevelFatal:
		return logrus.FatalLevel
	default:
		return logrus.InfoLevel
	}
}

// buildFormatter creates a logrus formatter based on config.
func buildFormatter(cfg *config) logrus.Formatter {
	callerPrettyfier := func(frame *runtime.Frame) (function string, file string) {
		filename := filepath.Base(frame.File)
		return "", filename + ":" + strconv.Itoa(frame.Line)
	}

	switch cfg.formatter {
	case FormatterJSON:
		return &logrus.JSONFormatter{
			TimestampFormat:  cfg.timeLayout,
			CallerPrettyfier: callerPrettyfier,
		}
	case FormatterText:
		baseFormatter := &logrus.TextFormatter{
			FullTimestamp:    cfg.development,
			TimestampFormat:  cfg.timeLayout,
			CallerPrettyfier: callerPrettyfier,
		}
		if cfg.development {
			baseFormatter.ForceColors = true
		}
		return &fullLevelTextFormatter{TextFormatter: baseFormatter}
	default:
		baseFormatter := &logrus.TextFormatter{
			CallerPrettyfier: callerPrettyfier,
		}
		return &fullLevelTextFormatter{TextFormatter: baseFormatter}
	}
}
