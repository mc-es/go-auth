package driver

import "go-auth/pkg/logger/internal/core"

type Logger interface {
	Debug(msg string, attrs ...core.Attr)
	Info(msg string, attrs ...core.Attr)
	Warn(msg string, attrs ...core.Attr)
	Error(msg string, attrs ...core.Attr)
	Panic(msg string, attrs ...core.Attr)
	Fatal(msg string, attrs ...core.Attr)

	With(attrs ...core.Attr) Logger
	Sync() error
}
