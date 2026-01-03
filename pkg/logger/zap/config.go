package zap

//nolint:revive
type (
	Level    int8
	Encoding string
	Option   func(*config)
)

type config struct {
	// core
	level       Level    // debug, info, warn, error, fatal
	encoding    Encoding // json, console
	outputPaths []string // log output targets
	development bool     // enables dev-friendly settings

	// custom settings
	timeLayout        string         // custom time layout
	initialFields     map[string]any // predefined structured fields
	disableCaller     bool           // disable caller info
	disableStacktrace bool           // disable stacktrace output

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
	defaultCallerSkip         = 2 // two levels: zap adapter -> logger package -> main package
	defaultSamplingInitial    = 100
	defaultSamplingThereafter = 100
)
