package zap

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go-auth/pkg/logger/internal/core"
)

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

func toZapFields(attrs []any) []any {
	dirty := false

	if len(attrs)%2 != 0 {
		dirty = true
	} else {
		for i := 0; i < len(attrs); i += 2 {
			key, ok := attrs[i].(string)
			if !ok {
				dirty = true

				break
			}

			if isReserved(key) {
				dirty = true

				break
			}
		}
	}

	if !dirty {
		return attrs
	}

	targetLen := len(attrs)
	if targetLen%2 != 0 {
		targetLen++
	}

	newAttrs := make([]any, 0, targetLen)

	for i := 0; i < len(attrs); i += 2 {
		key := normalizeKey(attrs[i])

		var val any = "_MISSING_"
		if i+1 < len(attrs) {
			val = attrs[i+1]
		}

		newAttrs = append(newAttrs, key, val)
	}

	return newAttrs
}

func normalizeKey(raw any) string {
	key, ok := raw.(string)
	if !ok {
		key = fmt.Sprint(raw)
	}

	if isReserved(key) {
		return "field." + key
	}

	return key
}

func isReserved(k string) bool {
	switch k {
	case "level", "msg", "time", "caller", "stacktrace":
		return true
	}

	return false
}
