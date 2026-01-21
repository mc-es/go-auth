package zerolog

import (
	"context"
	"io"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"go-auth/pkg/logger/internal/core"
	"go-auth/pkg/logger/internal/output"
	"go-auth/pkg/logger/internal/registry"
)

type adapter struct {
	logger    zerolog.Logger
	dests     *output.Destination
	extractor core.ExtractCtxFunc
	caller    bool
	name      string
}

//nolint:gochecknoinits
func init() {
	registry.Register(core.Driver("zerolog"), newZerolog)
}

func newZerolog(config *core.Config) (core.Logger, error) {
	dests, err := output.New(config)
	if err != nil {
		return nil, err
	}

	writers := buildWriters(dests, config)
	multiWriter := io.MultiWriter(writers...)

	zerolog.LevelFieldName = "level"
	zerolog.TimestampFieldName = "time"
	zerolog.MessageFieldName = "msg"
	zerolog.CallerFieldName = "caller"
	zerolog.ErrorStackFieldName = "stacktrace"
	zerolog.TimeFieldFormat = string(config.TimeLayout)

	loggerCtx := zerolog.New(multiWriter).With().Timestamp()
	if config.Development {
		loggerCtx = loggerCtx.Caller()
	}

	zlog := loggerCtx.Logger().Level(toZerologLevel(config.Level))
	log.Logger = zlog

	return &adapter{
		logger:    zlog,
		dests:     dests,
		extractor: config.Extractor,
		caller:    config.Development,
	}, nil
}

func (a *adapter) Debug(msg string, attrs ...any) {
	a.log(zerolog.DebugLevel, msg, attrs)
}

func (a *adapter) Info(msg string, attrs ...any) {
	a.log(zerolog.InfoLevel, msg, attrs)
}

func (a *adapter) Warn(msg string, attrs ...any) {
	a.log(zerolog.WarnLevel, msg, attrs)
}

func (a *adapter) Error(msg string, attrs ...any) {
	a.log(zerolog.ErrorLevel, msg, attrs)
}

func (a *adapter) Panic(msg string, attrs ...any) {
	a.log(zerolog.PanicLevel, msg, attrs)
}

func (a *adapter) Fatal(msg string, attrs ...any) {
	a.log(zerolog.FatalLevel, msg, attrs)
}

func (a *adapter) DebugCtx(ctx context.Context, msg string, attrs ...any) {
	a.logWithCtx(ctx, zerolog.DebugLevel, msg, attrs)
}

func (a *adapter) InfoCtx(ctx context.Context, msg string, attrs ...any) {
	a.logWithCtx(ctx, zerolog.InfoLevel, msg, attrs)
}

func (a *adapter) WarnCtx(ctx context.Context, msg string, attrs ...any) {
	a.logWithCtx(ctx, zerolog.WarnLevel, msg, attrs)
}

func (a *adapter) ErrorCtx(ctx context.Context, msg string, attrs ...any) {
	a.logWithCtx(ctx, zerolog.ErrorLevel, msg, attrs)
}

func (a *adapter) PanicCtx(ctx context.Context, msg string, attrs ...any) {
	a.logWithCtx(ctx, zerolog.PanicLevel, msg, attrs)
}

func (a *adapter) FatalCtx(ctx context.Context, msg string, attrs ...any) {
	a.logWithCtx(ctx, zerolog.FatalLevel, msg, attrs)
}

func (a *adapter) Named(name string) core.Logger {
	newName := name
	if a.name != "" {
		newName = a.name + "." + name
	}

	return &adapter{
		logger:    a.logger,
		dests:     a.dests,
		extractor: a.extractor,
		caller:    a.caller,
		name:      newName,
	}
}

func (a *adapter) Sync() error {
	if a.dests == nil {
		return nil
	}

	return a.dests.Close()
}

func (a *adapter) log(level zerolog.Level, msg string, attrs []any) {
	a.buildEvent(level, attrs).Msg(msg)
}

func (a *adapter) logWithCtx(ctx context.Context, level zerolog.Level, msg string, attrs []any) {
	if a.extractor != nil && ctx != nil {
		extracted := a.extractor(ctx)
		if len(extracted) > 0 {
			attrs = append(extracted, attrs...)
		}
	}

	a.buildEvent(level, attrs).Msg(msg)
}

func (a *adapter) buildEvent(level zerolog.Level, attrs []any) *zerolog.Event {
	event := a.logger.WithLevel(level)
	if a.caller {
		event = event.CallerSkipFrame(2)
	}

	for i := 0; i < len(attrs)-1; i += 2 {
		key := attrs[i].(string)
		value := attrs[i+1]
		event = addField(event, key, value)
	}

	if a.name != "" {
		event = event.Str("logger", a.name)
	}

	return event
}

func buildWriters(dests *output.Destination, config *core.Config) []io.Writer {
	var writers []io.Writer

	if dests.Console != nil {
		if config.Format == core.FormatText {
			writers = append(writers, zerolog.ConsoleWriter{
				Out:        dests.Console,
				TimeFormat: string(config.TimeLayout),
				NoColor:    !config.Development,
			})
		} else {
			writers = append(writers, dests.Console)
		}
	}

	if dests.File != nil {
		writers = append(writers, dests.File)
	}

	return writers
}

func toZerologLevel(l core.Level) zerolog.Level {
	switch l {
	case core.LevelDebug:
		return zerolog.DebugLevel
	case core.LevelInfo:
		return zerolog.InfoLevel
	case core.LevelWarn:
		return zerolog.WarnLevel
	case core.LevelError:
		return zerolog.ErrorLevel
	case core.LevelPanic:
		return zerolog.PanicLevel
	case core.LevelFatal:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

func addField(event *zerolog.Event, key string, value any) *zerolog.Event {
	switch val := value.(type) {
	case string:
		return event.Str(key, val)
	case int:
		return event.Int(key, val)
	case int8:
		return event.Int8(key, val)
	case float32:
		return event.Float32(key, val)
	case bool:
		return event.Bool(key, val)
	case time.Time:
		return event.Time(key, val)
	case time.Duration:
		return event.Dur(key, val)
	case error:
		return event.AnErr(key, val)
	default:
		return event.Interface(key, val)
	}
}
