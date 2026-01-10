package zap

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go-auth/pkg/logger/internal/core"
	"go-auth/pkg/logger/internal/driver"
	"go-auth/pkg/logger/internal/output"
)

type adapter struct {
	logger    *zap.SugaredLogger
	dests     *output.Destination
	extractor core.ExtractCtxFunc
}

//nolint:gochecknoinits
func init() {
	driver.Register(core.Driver("zap"), newZap)
}

func newZap(config *core.Config) (driver.Logger, error) {
	dests, err := output.New(config)
	if err != nil {
		return nil, err
	}

	var cores []zapcore.Core

	if dests.Console != nil {
		cores = append(cores, zapcore.NewCore(
			buildEncoder(config, true),
			zapcore.AddSync(dests.Console),
			toZapLevel(config.Level),
		))
	}

	if dests.File != nil {
		cores = append(cores, zapcore.NewCore(
			buildEncoder(config, false),
			zapcore.AddSync(dests.File),
			toZapLevel(config.Level),
		))
	}

	teeCore := zapcore.NewTee(cores...)
	z := zap.New(teeCore, buildOptions(config)...)

	return &adapter{
		logger:    z.Sugar(),
		dests:     dests,
		extractor: config.Extractor,
	}, nil
}

func (a *adapter) Debug(msg string, attrs ...any) {
	a.logger.Debugw(msg, toZapFields(attrs)...)
}

func (a *adapter) Info(msg string, attrs ...any) {
	a.logger.Infow(msg, toZapFields(attrs)...)
}

func (a *adapter) Warn(msg string, attrs ...any) {
	a.logger.Warnw(msg, toZapFields(attrs)...)
}

func (a *adapter) Error(msg string, attrs ...any) {
	a.logger.Errorw(msg, toZapFields(attrs)...)
}

func (a *adapter) Panic(msg string, attrs ...any) {
	a.logger.Panicw(msg, toZapFields(attrs)...)
}

func (a *adapter) Fatal(msg string, attrs ...any) {
	a.logger.Fatalw(msg, toZapFields(attrs)...)
}

func (a *adapter) DebugCtx(ctx context.Context, msg string, attrs ...any) {
	a.logger.Debugw(msg, toZapFields(a.extract(ctx, attrs))...)
}

func (a *adapter) InfoCtx(ctx context.Context, msg string, attrs ...any) {
	a.logger.Infow(msg, toZapFields(a.extract(ctx, attrs))...)
}

func (a *adapter) WarnCtx(ctx context.Context, msg string, attrs ...any) {
	a.logger.Warnw(msg, toZapFields(a.extract(ctx, attrs))...)
}

func (a *adapter) ErrorCtx(ctx context.Context, msg string, attrs ...any) {
	a.logger.Errorw(msg, toZapFields(a.extract(ctx, attrs))...)
}

func (a *adapter) PanicCtx(ctx context.Context, msg string, attrs ...any) {
	a.logger.Panicw(msg, toZapFields(a.extract(ctx, attrs))...)
}

func (a *adapter) FatalCtx(ctx context.Context, msg string, attrs ...any) {
	a.logger.Fatalw(msg, toZapFields(a.extract(ctx, attrs))...)
}

func (a *adapter) Sync() error {
	var err error
	if syncErr := a.logger.Sync(); syncErr != nil {
		err = syncErr
	}

	if a.dests != nil {
		err = errors.Join(err, a.dests.Close())
	}

	return err
}

func (a *adapter) extract(ctx context.Context, attrs []any) []any {
	if a.extractor != nil && ctx != nil {
		extracted := a.extractor(ctx)
		if len(extracted) > 0 {
			return append(extracted, attrs...)
		}
	}

	return attrs
}
