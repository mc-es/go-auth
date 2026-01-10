package logrus

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"go-auth/pkg/logger/internal/core"
)

func makeBaseLogger(cfg *core.Config) *logrus.Logger {
	l := logrus.New()
	l.SetLevel(toLogrusLevel(cfg.Level))

	return l
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

func toLogrusFields(attrs []any) logrus.Fields {
	fields := make(logrus.Fields, (len(attrs)+1)/2)

	for i := 0; i < len(attrs); i += 2 {
		key := normalizeKey(attrs[i])

		var val any = "_MISSING_"
		if i+1 < len(attrs) {
			val = attrs[i+1]
		}

		fields[key] = val
	}

	return fields
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
	case "level", "msg", "time", "caller", "logrus_error":
		return true
	}

	return false
}
