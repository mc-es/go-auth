package logrus

import (
	"io"
	"os"
	"time"
)

//nolint:revive
type (
	Level     int8
	Formatter string
	Option    func(*config)
)

type config struct {
	// core
	level       Level     // debug, info, warn, error, fatal
	formatter   Formatter // text, json
	output      io.Writer // log output target
	development bool      // enables dev-friendly settings

	// custom settings
	timeLayout        string         // custom time layout
	initialFields     map[string]any // predefined structured fields
	disableCaller     bool           // disable caller info
	disableStacktrace bool           // disable stacktrace output
}

//nolint:revive
const (
	LevelDebug Level = -1
	LevelInfo  Level = 0
	LevelWarn  Level = 1
	LevelError Level = 2
	LevelFatal Level = 3

	FormatterText Formatter = "text"
	FormatterJSON Formatter = "json"
)

const (
	defaultCallerSkip = 2
)

func defaultConfig() config {
	return config{
		level:       LevelDebug,
		formatter:   FormatterJSON,
		output:      os.Stdout,
		development: false,

		timeLayout:        time.DateTime,
		initialFields:     make(map[string]any),
		disableCaller:     false,
		disableStacktrace: false,
	}
}

func buildConfig(opts ...Option) config {
	cfg := defaultConfig()

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		opt(&cfg)
	}

	return cfg
}
