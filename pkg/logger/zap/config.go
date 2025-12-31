package zap

import "go.uber.org/zap/zapcore"

//nolint:revive
type (
	Level    int8
	Encoding string
	Option   func(*config) error
)

type config struct {
	// core
	level       Level    // debug, info, warn, error, fatal
	encoding    Encoding // json, console
	development bool     // enables dev-friendly settings

	// outputs
	outputPaths      []string // log output targets
	errorOutputPaths []string // error output targets

	// formatting
	timeEncoder   zapcore.TimeEncoder // custom time encoder
	initialFields map[string]any      // predefined structured fields

	// caller
	disableCaller bool // disable caller info

	// stacktrace
	disableStacktrace bool  // disable stacktrace output
	stacktraceLevel   Level // minimum level that triggers a stacktrace

	// sampling
	sampling           bool // enable log sampling
	samplingInitial    int  // log every message until this count
	samplingThereafter int  // log every Nth message afterward
}

//nolint:revive
const (
	LevelDebug Level = -1
	LevelInfo  Level = 0
	LevelWarn  Level = 1
	LevelError Level = 2
	LevelFatal Level = 3

	EncodingJson    Encoding = "json"
	EncodingConsole Encoding = "console"
)

const (
	defaultCallerSkip         = 2
	defaultSamplingInitial    = 100
	defaultSamplingThereafter = 100
)
