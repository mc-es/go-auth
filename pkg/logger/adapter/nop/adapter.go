package nop

import (
	"context"

	"go-auth/pkg/logger/internal/core"
	"go-auth/pkg/logger/internal/provider"
)

// adapter is a no-operation logger that discards all log messages.
type adapter struct{}

//nolint:gochecknoinits
func init() {
	provider.Register(core.Driver("nop"), newNop)
}

func newNop(_ *core.Config) (provider.Logger, error) {
	return &adapter{}, nil
}

func (a *adapter) Debug(_ string, _ ...any) { /* no-op */ }
func (a *adapter) Info(_ string, _ ...any)  { /* no-op */ }
func (a *adapter) Warn(_ string, _ ...any)  { /* no-op */ }
func (a *adapter) Error(_ string, _ ...any) { /* no-op */ }
func (a *adapter) Panic(_ string, _ ...any) { /* no-op */ }
func (a *adapter) Fatal(_ string, _ ...any) { /* no-op */ }

func (a *adapter) DebugCtx(_ context.Context, _ string, _ ...any) { /* no-op */ }
func (a *adapter) InfoCtx(_ context.Context, _ string, _ ...any)  { /* no-op */ }
func (a *adapter) WarnCtx(_ context.Context, _ string, _ ...any)  { /* no-op */ }
func (a *adapter) ErrorCtx(_ context.Context, _ string, _ ...any) { /* no-op */ }
func (a *adapter) PanicCtx(_ context.Context, _ string, _ ...any) { /* no-op */ }
func (a *adapter) FatalCtx(_ context.Context, _ string, _ ...any) { /* no-op */ }

func (a *adapter) Named(_ string) provider.Logger { return a }
func (a *adapter) Sync() error                    { return nil }
