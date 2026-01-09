package core

import (
	"context"
	"time"
)

type (
	Driver         string
	Level          int8
	Format         string
	TimeLayout     string
	ExtractCtxFunc func(ctx context.Context) []Attr
)

type FileRotation struct {
	MaxAge     int  // days
	MaxSize    int  // MB
	MaxBackups int  // count
	LocalTime  bool // local time in the backup file name
	Compress   bool
}

type Config struct {
	Driver       Driver
	Level        Level
	Format       Format
	TimeLayout   TimeLayout
	OutputPaths  []string
	Development  bool
	FileRotation FileRotation
	Extractor    ExtractCtxFunc
}

const (
	LevelDebug Level = iota - 1
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic
	LevelFatal
)

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

const (
	TimeLayoutDateTime TimeLayout = time.DateTime
	TimeLayoutDateOnly TimeLayout = time.DateOnly
	TimeLayoutTimeOnly TimeLayout = time.TimeOnly
	TimeLayoutRFC3339  TimeLayout = time.RFC3339
	TimeLayoutRFC822   TimeLayout = time.RFC822
	TimeLayoutRFC1123  TimeLayout = time.RFC1123
)
