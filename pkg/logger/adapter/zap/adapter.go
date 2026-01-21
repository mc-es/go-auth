package zap

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go-auth/pkg/logger/internal/core"
	"go-auth/pkg/logger/internal/output"
	"go-auth/pkg/logger/internal/registry"
)

type adapter struct {
	logger    *zap.SugaredLogger
	dests     *output.Destination
	extractor core.ExtractCtxFunc
}

//nolint:gochecknoinits
func init() {
	registry.Register(core.Driver("zap"), newZap)
}

func newZap(config *core.Config) (core.Logger, error) {
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
	a.log(zapcore.DebugLevel, msg, attrs)
}

func (a *adapter) Info(msg string, attrs ...any) {
	a.log(zapcore.InfoLevel, msg, attrs)
}

func (a *adapter) Warn(msg string, attrs ...any) {
	a.log(zapcore.WarnLevel, msg, attrs)
}

func (a *adapter) Error(msg string, attrs ...any) {
	a.log(zapcore.ErrorLevel, msg, attrs)
}

func (a *adapter) Panic(msg string, attrs ...any) {
	a.log(zapcore.PanicLevel, msg, attrs)
}

func (a *adapter) Fatal(msg string, attrs ...any) {
	a.log(zapcore.FatalLevel, msg, attrs)
}

func (a *adapter) DebugCtx(ctx context.Context, msg string, attrs ...any) {
	a.logWithCtx(ctx, zapcore.DebugLevel, msg, attrs)
}

func (a *adapter) InfoCtx(ctx context.Context, msg string, attrs ...any) {
	a.logWithCtx(ctx, zapcore.InfoLevel, msg, attrs)
}

func (a *adapter) WarnCtx(ctx context.Context, msg string, attrs ...any) {
	a.logWithCtx(ctx, zapcore.WarnLevel, msg, attrs)
}

func (a *adapter) ErrorCtx(ctx context.Context, msg string, attrs ...any) {
	a.logWithCtx(ctx, zapcore.ErrorLevel, msg, attrs)
}

func (a *adapter) PanicCtx(ctx context.Context, msg string, attrs ...any) {
	a.logWithCtx(ctx, zapcore.PanicLevel, msg, attrs)
}

func (a *adapter) FatalCtx(ctx context.Context, msg string, attrs ...any) {
	a.logWithCtx(ctx, zapcore.FatalLevel, msg, attrs)
}

func (a *adapter) Named(name string) core.Logger {
	return &adapter{
		logger:    a.logger.Named(name),
		dests:     a.dests,
		extractor: a.extractor,
	}
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

func (a *adapter) log(lvl zapcore.Level, msg string, attrs []any) {
	a.logger.Logw(lvl, msg, attrs...)
}

func (a *adapter) logWithCtx(ctx context.Context, lvl zapcore.Level, msg string, attrs []any) {
	if a.extractor != nil && ctx != nil {
		extracted := a.extractor(ctx)
		if len(extracted) > 0 {
			attrs = append(extracted, attrs...)
		}
	}

	a.logger.Logw(lvl, msg, attrs...)
}

func buildEncoder(cfg *core.Config, isConsole bool) zapcore.Encoder {
	enCfg := zap.NewProductionEncoderConfig()

	if cfg.Development {
		enCfg = zap.NewDevelopmentEncoderConfig()
	}

	enCfg.NameKey = "logger"
	enCfg.LevelKey = "level"
	enCfg.TimeKey = "time"
	enCfg.MessageKey = "msg"
	enCfg.CallerKey = "caller"
	enCfg.StacktraceKey = "stacktrace"

	enCfg.EncodeTime = zapcore.TimeEncoderOfLayout(string(cfg.TimeLayout))
	enCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	if cfg.Format == core.FormatText && isConsole && cfg.Development {
		enCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if cfg.Format == core.FormatJSON {
		return zapcore.NewJSONEncoder(enCfg)
	}

	return zapcore.NewConsoleEncoder(enCfg)
}

func toZapLevel(l core.Level) zapcore.Level {
	switch l {
	case core.LevelDebug:
		return zapcore.DebugLevel
	case core.LevelInfo:
		return zapcore.InfoLevel
	case core.LevelWarn:
		return zapcore.WarnLevel
	case core.LevelError:
		return zapcore.ErrorLevel
	case core.LevelPanic:
		return zapcore.PanicLevel
	case core.LevelFatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func buildOptions(cfg *core.Config) []zap.Option {
	opts := []zap.Option{
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	if cfg.Development {
		opts = append(opts, zap.WithCaller(true), zap.AddCallerSkip(2))
	}

	return opts
}
