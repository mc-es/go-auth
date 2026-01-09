package driver

import (
	"context"

	"go-auth/pkg/logger/internal/core"
)

type Logger interface {
	Debug(msg string, attrs ...core.Attr)
	Info(msg string, attrs ...core.Attr)
	Warn(msg string, attrs ...core.Attr)
	Error(msg string, attrs ...core.Attr)
	Panic(msg string, attrs ...core.Attr)
	Fatal(msg string, attrs ...core.Attr)

	DebugCtx(ctx context.Context, msg string, attrs ...core.Attr)
	InfoCtx(ctx context.Context, msg string, attrs ...core.Attr)
	WarnCtx(ctx context.Context, msg string, attrs ...core.Attr)
	ErrorCtx(ctx context.Context, msg string, attrs ...core.Attr)
	PanicCtx(ctx context.Context, msg string, attrs ...core.Attr)
	FatalCtx(ctx context.Context, msg string, attrs ...core.Attr)

	Sync() error
}
