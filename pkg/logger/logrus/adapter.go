// Package logrus is the adapter for the logrus logger.
package logrus

import (
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"

	"go-auth/pkg/logger"
)

type adapter struct {
	loggers    []*logrus.Logger
	fields     logrus.Fields
	showCaller bool
}

// Register registers the logrus adapter.
func Register() {
	logger.Register(logger.DriverLogrus, newLogrus)
}

func newLogrus(config *logger.Config) logger.Logger {
	var loggers []*logrus.Logger

	consolePaths, filePaths := splitPaths(config.OutputPath)

	if len(consolePaths) > 0 {
		l := makeBaseLogger(config)
		l.SetOutput(buildOutput(consolePaths))
		setupFormatter(l, config, true)
		loggers = append(loggers, l)
	}

	if len(filePaths) > 0 {
		l := makeBaseLogger(config)
		l.SetOutput(buildOutput(filePaths))
		setupFormatter(l, config, false)
		loggers = append(loggers, l)
	}

	return &adapter{
		loggers:    loggers,
		fields:     make(logrus.Fields),
		showCaller: !config.DisableCaller,
	}
}

// Debug logs a debug message.
func (a *adapter) Debug(msg string, args ...any) { a.dispatch(logrus.DebugLevel, msg, args...) }

// Info logs an info message.
func (a *adapter) Info(msg string, args ...any) { a.dispatch(logrus.InfoLevel, msg, args...) }

// Warn logs a warning message.
func (a *adapter) Warn(msg string, args ...any) { a.dispatch(logrus.WarnLevel, msg, args...) }

// Error logs an error message.
func (a *adapter) Error(msg string, args ...any) { a.dispatch(logrus.ErrorLevel, msg, args...) }

// Panic logs a panic message.
func (a *adapter) Panic(msg string, args ...any) { a.dispatch(logrus.PanicLevel, msg, args...) }

// Fatal logs a fatal message.
func (a *adapter) Fatal(msg string, args ...any) { a.dispatch(logrus.FatalLevel, msg, args...) }

// With returns a logger with additional fields.
func (a *adapter) With(args ...any) logger.Logger {
	newFields := make(logrus.Fields, len(a.fields)+len(args)/2)
	maps.Copy(newFields, a.fields)

	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			newFields[key] = args[i+1]
		}
	}

	return &adapter{
		loggers:    a.loggers,
		fields:     newFields,
		showCaller: a.showCaller,
	}
}

// Sync flushes buffered log entries.
func (a *adapter) Sync() error { return nil }

func (a *adapter) dispatch(level logrus.Level, msg string, args ...any) {
	for _, l := range a.loggers {
		entry := l.WithFields(a.fields)

		if len(args) > 0 {
			tempFields := make(logrus.Fields)

			for i := 0; i < len(args)-1; i += 2 {
				if key, ok := args[i].(string); ok {
					tempFields[key] = args[i+1]
				}
			}

			entry = entry.WithFields(tempFields)
		}

		if a.showCaller {
			if _, file, line, ok := runtime.Caller(2); ok {
				caller := fmt.Sprintf("%s:%d", filepath.Base(file), line)
				entry = entry.WithField("caller", caller)
			}
		}

		entry.Log(level, msg)
	}
}

func makeBaseLogger(cfg *logger.Config) *logrus.Logger {
	l := logrus.New()
	l.SetLevel(mapLevel(int8(cfg.Level)))
	l.SetReportCaller(false)

	return l
}

func mapLevel(lvl int8) logrus.Level {
	switch lvl {
	case int8(logger.LevelDebug):
		return logrus.DebugLevel
	case int8(logger.LevelInfo):
		return logrus.InfoLevel
	case int8(logger.LevelWarn):
		return logrus.WarnLevel
	case int8(logger.LevelError):
		return logrus.ErrorLevel
	case int8(logger.LevelFatal):
		return logrus.FatalLevel
	case int8(logger.LevelPanic):
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

func setupFormatter(log *logrus.Logger, cfg *logger.Config, allowColor bool) {
	if cfg.Formatter == logger.FormatterJSON {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: string(cfg.TimeLayout),
			PrettyPrint:     allowColor && cfg.DevMode,
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: string(cfg.TimeLayout),
			FullTimestamp:   true,
			ForceColors:     allowColor && cfg.DevMode,
			DisableColors:   !allowColor,
		})
	}
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

func buildOutput(paths []string) io.Writer {
	writers := make([]io.Writer, 0, len(paths))

	for _, path := range paths {
		if path == "stdout" {
			writers = append(writers, os.Stdout)

			continue
		}

		if path == "stderr" {
			writers = append(writers, os.Stderr)

			continue
		}

		_ = os.MkdirAll(filepath.Dir(path), 0o750)

		file, err := os.OpenFile(filepath.Clean(path), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			panic(err)
		}

		writers = append(writers, file)
	}

	if len(writers) == 1 {
		return writers[0]
	}

	return io.MultiWriter(writers...)
}
