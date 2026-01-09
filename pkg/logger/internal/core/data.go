package core

import "time"

type Driver string

type Level int8

const (
	LevelDebug Level = iota - 1
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic
	LevelFatal
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

type TimeLayout string

const (
	TimeLayoutDateTime TimeLayout = time.DateTime
	TimeLayoutDateOnly TimeLayout = time.DateOnly
	TimeLayoutTimeOnly TimeLayout = time.TimeOnly
	TimeLayoutRFC3339  TimeLayout = time.RFC3339
	TimeLayoutRFC822   TimeLayout = time.RFC822
	TimeLayoutRFC1123  TimeLayout = time.RFC1123
)

type Config struct {
	Driver      Driver
	Level       Level
	Format      Format
	TimeLayout  TimeLayout
	OutputPaths []string
	Development bool
}
