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
	a.log(context.TODO(), logrus.DebugLevel, msg, attrs)
}

func (a *adapter) Info(msg string, attrs ...any) {
	a.log(context.TODO(), logrus.InfoLevel, msg, attrs)
}

func (a *adapter) Warn(msg string, attrs ...any) {
	a.log(context.TODO(), logrus.WarnLevel, msg, attrs)
}

func (a *adapter) Error(msg string, attrs ...any) {
	a.log(context.TODO(), logrus.ErrorLevel, msg, attrs)
}

func (a *adapter) Panic(msg string, attrs ...any) {
	a.log(context.TODO(), logrus.PanicLevel, msg, attrs)
}

func (a *adapter) Fatal(msg string, attrs ...any) {
	a.log(context.TODO(), logrus.FatalLevel, msg, attrs)
}

func (a *adapter) DebugCtx(ctx context.Context, msg string, attrs ...any) {
	a.log(ctx, logrus.DebugLevel, msg, attrs)
}

func (a *adapter) InfoCtx(ctx context.Context, msg string, attrs ...any) {
	a.log(ctx, logrus.InfoLevel, msg, attrs)
}

func (a *adapter) WarnCtx(ctx context.Context, msg string, attrs ...any) {
	a.log(ctx, logrus.WarnLevel, msg, attrs)
}

func (a *adapter) ErrorCtx(ctx context.Context, msg string, attrs ...any) {
	a.log(ctx, logrus.ErrorLevel, msg, attrs)
}

func (a *adapter) PanicCtx(ctx context.Context, msg string, attrs ...any) {
	a.log(ctx, logrus.PanicLevel, msg, attrs)
}

func (a *adapter) FatalCtx(ctx context.Context, msg string, attrs ...any) {
	a.log(ctx, logrus.FatalLevel, msg, attrs)
}

func (a *adapter) Sync() error {
	if a.dests == nil {
		return nil
	}

	return a.dests.Close()
}

func (a *adapter) log(ctx context.Context, level logrus.Level, msg string, attrs []any) {
	if a.extractor != nil && ctx != nil {
		extracted := a.extractor(ctx)
		if len(extracted) > 0 {
			attrs = append(extracted, attrs...)
		}
	}

	fields := make(logrus.Fields, len(attrs)/2)
	for i := 0; i < len(attrs); i += 2 {
		key := attrs[i].(string)
		fields[key] = attrs[i+1]
	}

	if a.caller {
		if _, file, line, ok := runtime.Caller(2); ok {
			shortFile := path.Base(file) + ":" + strconv.Itoa(line)
			fields["caller"] = shortFile
		}
	}

	for _, l := range a.loggers {
		l.WithFields(fields).Log(level, msg)
	}
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
