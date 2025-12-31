// Package logger provides a logging interface.
package logger

// Logger defines the logging interface.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)
	With(args ...any) Logger
	// WithContext(ctx context.Context) Logger // TODO: implement context support
	Sync() error
}

var globalLogger Logger

func get() Logger {
	if globalLogger == nil {
		panic("global logger is not initialized. Call SetGlobalLogger() first.")
	}

	return globalLogger
}

// SetGlobalLogger sets the global logger instance.
func SetGlobalLogger(logger Logger) { globalLogger = logger }

// Debug logs a debug message.
func Debug(msg string, args ...any) { get().Debug(msg, args...) }

// Info logs an info message.
func Info(msg string, args ...any) { get().Info(msg, args...) }

// Warn logs a warning message.
func Warn(msg string, args ...any) { get().Warn(msg, args...) }

// Error logs an error message.
func Error(msg string, args ...any) { get().Error(msg, args...) }

// Fatal logs a fatal message and exits.
func Fatal(msg string, args ...any) { get().Fatal(msg, args...) }

// With returns a logger with additional fields.
func With(args ...any) Logger { return get().With(args...) }

// Sync flushes buffered log entries.
func Sync() error { return get().Sync() }
