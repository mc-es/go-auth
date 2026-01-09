package zap

import (
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go-auth/pkg/logger/internal/core"
	"go-auth/pkg/logger/internal/driver"
	"go-auth/pkg/logger/internal/output"
)

type adapter struct {
	logger *zap.SugaredLogger
	dests  *output.Destination
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
		logger: z.Sugar(),
		dests:  dests,
	}, nil
}

func (a *adapter) Debug(msg string, attrs ...core.Attr) {
	a.logger.Debugw(msg, toZapArgs(attrs)...)
}

func (a *adapter) Info(msg string, attrs ...core.Attr) {
	a.logger.Infow(msg, toZapArgs(attrs)...)
}

func (a *adapter) Warn(msg string, attrs ...core.Attr) {
	a.logger.Warnw(msg, toZapArgs(attrs)...)
}

func (a *adapter) Error(msg string, attrs ...core.Attr) {
	a.logger.Errorw(msg, toZapArgs(attrs)...)
}

func (a *adapter) Panic(msg string, attrs ...core.Attr) {
	a.logger.Panicw(msg, toZapArgs(attrs)...)
}

func (a *adapter) Fatal(msg string, attrs ...core.Attr) {
	a.logger.Fatalw(msg, toZapArgs(attrs)...)
}

func (a *adapter) With(attrs ...core.Attr) driver.Logger {
	return &adapter{
		logger: a.logger.With(toZapArgs(attrs)...),
		dests:  nil,
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

func toZapLevel(l core.Level) zapcore.Level {
	switch l {
	case core.LevelDebug, core.LevelInfo, core.LevelWarn, core.LevelError:
		return zapcore.Level(l)
	case core.LevelPanic:
		return zapcore.PanicLevel
	case core.LevelFatal:
		return zapcore.FatalLevel
	default:
		return zapcore.DebugLevel
	}
}

func toZapArgs(attrs []core.Attr) []any {
	args := make([]any, 0, len(attrs)*2)
	for _, a := range attrs {
		args = append(args, core.SanitizeKey(a.Key), a.Value)
	}

	return args
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

func buildOptions(cfg *core.Config) []zap.Option {
	opts := []zap.Option{
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	if cfg.Development {
		opts = append(opts, zap.WithCaller(true), zap.AddCallerSkip(1))
	}

	return opts
}
