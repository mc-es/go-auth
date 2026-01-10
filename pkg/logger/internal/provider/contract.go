package provider

import (
	"context"
)

type Logger interface {
	Debug(msg string, attrs ...any)
	Info(msg string, attrs ...any)
	Warn(msg string, attrs ...any)
	Error(msg string, attrs ...any)
	Panic(msg string, attrs ...any)
	Fatal(msg string, attrs ...any)

	DebugCtx(ctx context.Context, msg string, attrs ...any)
	InfoCtx(ctx context.Context, msg string, attrs ...any)
	WarnCtx(ctx context.Context, msg string, attrs ...any)
	ErrorCtx(ctx context.Context, msg string, attrs ...any)
	PanicCtx(ctx context.Context, msg string, attrs ...any)
	FatalCtx(ctx context.Context, msg string, attrs ...any)

	Sync() error
}
