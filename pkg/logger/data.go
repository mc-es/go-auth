package logger

import "time"

type driver string

// Drivers.
const (
	DriverZap    driver = "zap"
	DriverLogrus driver = "logrus"
)

type level int8

// Levels.
const (
	LevelDebug level = iota - 1
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelPanic
)

type formatter string

// Formatters.
const (
	FormatterJSON formatter = "json"
	FormatterText formatter = "text"
)

type timeLayout string

// Time layouts.
const (
	TimeLayoutDateTime timeLayout = time.DateTime
	TimeLayoutDateOnly timeLayout = time.DateOnly
	TimeLayoutTimeOnly timeLayout = time.TimeOnly
	TimeLayoutRFC3339  timeLayout = time.RFC3339
	TimeLayoutRFC822   timeLayout = time.RFC822
	TimeLayoutRFC1123  timeLayout = time.RFC1123
)

// Config represents the configuration of the logger.
type Config struct {
	Driver        driver
	Level         level
	Formatter     formatter
	TimeLayout    timeLayout
	OutputPath    []string
	DevMode       bool
	DisableCaller bool
}
