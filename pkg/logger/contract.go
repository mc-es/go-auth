// Package logger provides a logger interface and implementations for different logging drivers.
package logger

// Logger is the contract for the logger.
type Logger interface {
	Debug(message string, args ...any)
	Info(message string, args ...any)
	Warn(message string, args ...any)
	Error(message string, args ...any)
	Panic(message string, args ...any)
	Fatal(message string, args ...any)
	With(args ...any) Logger
	Sync() error
}
