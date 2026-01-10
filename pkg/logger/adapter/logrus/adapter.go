package logrus

import (
	"context"
	"io"
	"path"
	"runtime"
	"strconv"

	"github.com/sirupsen/logrus"

	"go-auth/pkg/logger/internal/core"
	"go-auth/pkg/logger/internal/output"
	"go-auth/pkg/logger/internal/provider"
)

type adapter struct {
	loggers   []*logrus.Logger
	dests     *output.Destination
	caller    bool
	extractor core.ExtractCtxFunc
}

//nolint:gochecknoinits
func init() {
	provider.Register(core.Driver("logrus"), newLogrus)
}

func newLogrus(config *core.Config) (provider.Logger, error) {
	dests, err := output.New(config)
	if err != nil {
		return nil, err
	}

	var loggers []*logrus.Logger

	if dests.Console != nil {
		consoleLogger := makeBaseLogger(config)
		consoleLogger.SetOutput(io.MultiWriter(dests.Console))
		setupFormatter(consoleLogger, config, true)
		loggers = append(loggers, consoleLogger)
	}

	if dests.File != nil {
		fileLogger := makeBaseLogger(config)
		fileLogger.SetOutput(io.MultiWriter(dests.File))
		setupFormatter(fileLogger, config, false)
		loggers = append(loggers, fileLogger)
	}

	return &adapter{
		loggers:   loggers,
		dests:     dests,
		caller:    config.Development,
		extractor: config.Extractor,
	}, nil
}

func (a *adapter) Debug(msg string, attrs ...any) {
	a.dispatch(logrus.DebugLevel, msg, attrs...)
}

func (a *adapter) Info(msg string, attrs ...any) {
	a.dispatch(logrus.InfoLevel, msg, attrs...)
}

func (a *adapter) Warn(msg string, attrs ...any) {
	a.dispatch(logrus.WarnLevel, msg, attrs...)
}

func (a *adapter) Error(msg string, attrs ...any) {
	a.dispatch(logrus.ErrorLevel, msg, attrs...)
}

func (a *adapter) Panic(msg string, attrs ...any) {
	a.dispatch(logrus.PanicLevel, msg, attrs...)
}

func (a *adapter) Fatal(msg string, attrs ...any) {
	a.dispatch(logrus.FatalLevel, msg, attrs...)
}

func (a *adapter) DebugCtx(ctx context.Context, msg string, attrs ...any) {
	a.dispatch(logrus.DebugLevel, msg, a.extract(ctx, attrs)...)
}

func (a *adapter) InfoCtx(ctx context.Context, msg string, attrs ...any) {
	a.dispatch(logrus.InfoLevel, msg, a.extract(ctx, attrs)...)
}

func (a *adapter) WarnCtx(ctx context.Context, msg string, attrs ...any) {
	a.dispatch(logrus.WarnLevel, msg, a.extract(ctx, attrs)...)
}

func (a *adapter) ErrorCtx(ctx context.Context, msg string, attrs ...any) {
	a.dispatch(logrus.ErrorLevel, msg, a.extract(ctx, attrs)...)
}

func (a *adapter) PanicCtx(ctx context.Context, msg string, attrs ...any) {
	a.dispatch(logrus.PanicLevel, msg, a.extract(ctx, attrs)...)
}

func (a *adapter) FatalCtx(ctx context.Context, msg string, attrs ...any) {
	a.dispatch(logrus.FatalLevel, msg, a.extract(ctx, attrs)...)
}

func (a *adapter) Sync() error {
	if a.dests == nil {
		return nil
	}

	return a.dests.Close()
}

func (a *adapter) dispatch(level logrus.Level, msg string, attrs ...any) {
	fields := toLogrusFields(attrs)

	if a.caller {
		if _, file, line, ok := runtime.Caller(2); ok {
			shortFile := path.Base(file) + ":" + strconv.Itoa(line)
			fields["caller"] = shortFile
		}
	}

	for _, l := range a.loggers {
		entry := l.WithFields(fields)
		entry.Log(level, msg)
	}
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
