package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the interface for structured logging operations.
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Sync() error
}

// SugarLogger is the interface for sugared (printf-style) logging operations.
type SugarLogger interface {
	// Formatted methods
	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	Fatalf(template string, args ...any)

	// Field-based methods
	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	Fatalw(msg string, keysAndValues ...any)
}

// Option mutates the logger configuration during construction.
type Option func(*config) error

// Types.
type (
	level    int8
	encoding string
)

// Logger implementation.
type zapLogger struct {
	*zap.Logger
}

// SugarLogger implementation.
type zapSugarLogger struct {
	*zap.SugaredLogger
}

// Config.
type config struct {
	// core
	level       level    // debug, info, warn, error, fatal
	encoding    encoding // json, console
	development bool     // enables dev-friendly settings

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
	disableStacktrace bool  // disable stacktrace output
	stacktraceLevel   level // minimum level that triggers a stacktrace

	// sampling
	sampling           bool // enable log sampling
	samplingInitial    int  // log every message until this count
	samplingThereafter int  // log every Nth message afterward
}

//nolint:revive
const (
	LevelDebug level = -1
	LevelInfo  level = 0
	LevelWarn  level = 1
	LevelError level = 2
	LevelFatal level = 3

	EncodingJson    encoding = "json"
	EncodingConsole encoding = "console"

	defaultSamplingInitial    = 100
	defaultSamplingThereafter = 100
)

// Global variables.
var (
	globalLogger Logger      // global logger instance
	globalSugar  SugarLogger // cached sugared logger
	initOnce     sync.Once   // ensures Init runs only once
	errInit      error       // stores Init error for repeated calls
)

// Compile-time assertions.
var (
	_ Logger      = (*zapLogger)(nil)
	_ SugarLogger = (*zapSugarLogger)(nil)
)
