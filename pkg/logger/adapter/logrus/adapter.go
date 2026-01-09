package logrus

import (
	"context"
	"io"
	"maps"
	"path"
	"runtime"
	"strconv"

	"github.com/sirupsen/logrus"

	"go-auth/pkg/logger/internal/core"
	"go-auth/pkg/logger/internal/driver"
	"go-auth/pkg/logger/internal/output"
)

type adapter struct {
	loggers   []*logrus.Logger
	fields    logrus.Fields
	dests     *output.Destination
	caller    bool
	extractor core.ExtractCtxFunc
}

//nolint:gochecknoinits
func init() {
	driver.Register(core.Driver("logrus"), newLogrus)
}

func newLogrus(config *core.Config) (driver.Logger, error) {
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
		fields:    make(logrus.Fields),
		dests:     dests,
		caller:    config.Development,
		extractor: config.Extractor,
	}, nil
}

func (a *adapter) Debug(msg string, attrs ...core.Attr) {
	a.dispatch(logrus.DebugLevel, msg, attrs...)
}

func (a *adapter) Info(msg string, attrs ...core.Attr) {
	a.dispatch(logrus.InfoLevel, msg, attrs...)
}

func (a *adapter) Warn(msg string, attrs ...core.Attr) {
	a.dispatch(logrus.WarnLevel, msg, attrs...)
}

func (a *adapter) Error(msg string, attrs ...core.Attr) {
	a.dispatch(logrus.ErrorLevel, msg, attrs...)
}

func (a *adapter) Panic(msg string, attrs ...core.Attr) {
	a.dispatch(logrus.PanicLevel, msg, attrs...)
}

func (a *adapter) Fatal(msg string, attrs ...core.Attr) {
	a.dispatch(logrus.FatalLevel, msg, attrs...)
}

func (a *adapter) DebugCtx(ctx context.Context, msg string, attrs ...core.Attr) {
	a.dispatch(logrus.DebugLevel, msg, a.extract(ctx, attrs)...)
}

func (a *adapter) InfoCtx(ctx context.Context, msg string, attrs ...core.Attr) {
	a.dispatch(logrus.InfoLevel, msg, a.extract(ctx, attrs)...)
}

func (a *adapter) WarnCtx(ctx context.Context, msg string, attrs ...core.Attr) {
	a.dispatch(logrus.WarnLevel, msg, a.extract(ctx, attrs)...)
}

func (a *adapter) ErrorCtx(ctx context.Context, msg string, attrs ...core.Attr) {
	a.dispatch(logrus.ErrorLevel, msg, a.extract(ctx, attrs)...)
}

func (a *adapter) PanicCtx(ctx context.Context, msg string, attrs ...core.Attr) {
	a.dispatch(logrus.PanicLevel, msg, a.extract(ctx, attrs)...)
}

func (a *adapter) FatalCtx(ctx context.Context, msg string, attrs ...core.Attr) {
	a.dispatch(logrus.FatalLevel, msg, a.extract(ctx, attrs)...)
}

func (a *adapter) Sync() error {
	if a.dests == nil {
		return nil
	}

	return a.dests.Close()
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

func makeBaseLogger(cfg *core.Config) *logrus.Logger {
	l := logrus.New()
	l.SetLevel(toLogrusLevel(cfg.Level))

	return l
}

func toLogrusLevel(l core.Level) logrus.Level {
	switch l {
	case core.LevelDebug:
		return logrus.DebugLevel
	case core.LevelInfo:
		return logrus.InfoLevel
	case core.LevelWarn:
		return logrus.WarnLevel
	case core.LevelError:
		return logrus.ErrorLevel
	case core.LevelPanic:
		return logrus.PanicLevel
	case core.LevelFatal:
		return logrus.FatalLevel
	default:
		return logrus.InfoLevel
	}
}

func setupFormatter(log *logrus.Logger, cfg *core.Config, allowColor bool) {
	if cfg.Format == core.FormatJSON {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: string(cfg.TimeLayout),
			PrettyPrint:     allowColor && cfg.Development,
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: string(cfg.TimeLayout),
			FullTimestamp:   true,
			ForceColors:     allowColor && cfg.Development,
			DisableColors:   !allowColor,
		})
	}
}

func (a *adapter) dispatch(level logrus.Level, msg string, attrs ...core.Attr) {
	fields := toLogrusFields(a.fields, attrs)

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

func toLogrusFields(base logrus.Fields, attrs []core.Attr) logrus.Fields {
	fields := make(logrus.Fields, len(base)+len(attrs))
	maps.Copy(fields, base)

	for _, a := range attrs {
		fields[core.SanitizeKey(a.Key)] = a.Value
	}

	return fields
}
