package logger

type Driver string

const (
	DriverZap    Driver = "zap"
	DriverLogrus Driver = "logrus"
	DriverNoOp   Driver = "noop"
)

type Level int8

const (
	LevelDebug Level = iota - 1
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

type Encoding string

const (
	EncodingJSON    Encoding = "json"
	EncodingConsole Encoding = "console"
)

type Config struct {
	Driver            Driver
	Level             Level
	Encoding          Encoding
	Development       bool
	DisableCaller     bool
	DisableStacktrace bool
	OutputPaths       []string
	InitialFields     map[string]any
	TimeLayout        string
	DriverOptions     map[string]any
}
