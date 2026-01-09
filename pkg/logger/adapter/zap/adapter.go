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

func (a *adapter) Debug(msg string, attrs ...core.Attr) {
	a.logger.Debugw(msg, toZapFields(attrs)...)
}

func (a *adapter) Info(msg string, attrs ...core.Attr) {
	a.logger.Infow(msg, toZapFields(attrs)...)
}

func (a *adapter) Warn(msg string, attrs ...core.Attr) {
	a.logger.Warnw(msg, toZapFields(attrs)...)
}

func (a *adapter) Error(msg string, attrs ...core.Attr) {
	a.logger.Errorw(msg, toZapFields(attrs)...)
}

func (a *adapter) Panic(msg string, attrs ...core.Attr) {
	a.logger.Panicw(msg, toZapFields(attrs)...)
}

func (a *adapter) Fatal(msg string, attrs ...core.Attr) {
	a.logger.Fatalw(msg, toZapFields(attrs)...)
}

func (a *adapter) DebugCtx(ctx context.Context, msg string, attrs ...core.Attr) {
	a.logger.Debugw(msg, toZapFields(a.extract(ctx, attrs))...)
}

func (a *adapter) InfoCtx(ctx context.Context, msg string, attrs ...core.Attr) {
	a.logger.Infow(msg, toZapFields(a.extract(ctx, attrs))...)
}

func (a *adapter) WarnCtx(ctx context.Context, msg string, attrs ...core.Attr) {
	a.logger.Warnw(msg, toZapFields(a.extract(ctx, attrs))...)
}

func (a *adapter) ErrorCtx(ctx context.Context, msg string, attrs ...core.Attr) {
	a.logger.Errorw(msg, toZapFields(a.extract(ctx, attrs))...)
}

func (a *adapter) PanicCtx(ctx context.Context, msg string, attrs ...core.Attr) {
	a.logger.Panicw(msg, toZapFields(a.extract(ctx, attrs))...)
}

func (a *adapter) FatalCtx(ctx context.Context, msg string, attrs ...core.Attr) {
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

func (a *adapter) extract(ctx context.Context, attrs []core.Attr) []core.Attr {
	if a.extractor != nil && ctx != nil {
		extracted := a.extractor(ctx)
		if len(extracted) > 0 {
			return append(extracted, attrs...)
		}
	}

	return attrs
}

func buildEncoder(cfg *core.Config, isConsole bool) zapcore.Encoder {
	enCfg := zap.NewProductionEncoderConfig()

	if cfg.Development {
		enCfg = zap.NewDevelopmentEncoderConfig()
	}

	enCfg.LevelKey = "level"
	enCfg.MessageKey = "msg"
	enCfg.TimeKey = "time"
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
	case core.LevelDebug, core.LevelInfo, core.LevelWarn, core.LevelError:
		return zapcore.Level(l)
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
		opts = append(opts, zap.WithCaller(true), zap.AddCallerSkip(1))
	}

	return opts
}

func toZapFields(attrs []core.Attr) []any {
	fields := make([]any, 0, len(attrs)*2)
	for _, a := range attrs {
		fields = append(fields, core.SanitizeKey(a.Key), a.Value)
	}

	return fields
}
