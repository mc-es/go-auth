package bootstrap

import (
	"strings"

	"go-auth/internal/config"
	"go-auth/pkg/logger"
)

func BuildLoggerOptions(cfg *config.Config) []logger.Option {
	loggerCfg := &cfg.Logger

	var opts []logger.Option

	if opt := driverOption(loggerCfg); opt != nil {
		opts = append(opts, opt)
	}

	if opt := levelOption(loggerCfg); opt != nil {
		opts = append(opts, opt)
	}

	if opt := formatOption(loggerCfg); opt != nil {
		opts = append(opts, opt)
	}

	if opt := timeLayoutOption(loggerCfg); opt != nil {
		opts = append(opts, opt)
	}

	if len(loggerCfg.OutputPaths) > 0 {
		opts = append(opts, logger.WithOutputPaths(loggerCfg.OutputPaths...))
	}

	if opt := fileRotationOption(loggerCfg); opt != nil {
		opts = append(opts, opt)
	}

	if loggerCfg.Development {
		opts = append(opts, logger.WithDevelopment())
	}

	return opts
}

func driverOption(cfg *config.Logger) logger.Option {
	switch strings.ToLower(cfg.Driver) {
	case "logrus":
		return logger.WithDriver(logger.DriverLogrus)
	case "zap":
		return logger.WithDriver(logger.DriverZap)
	case "zerolog":
		return logger.WithDriver(logger.DriverZerolog)
	default:
		return nil
	}
}

func levelOption(cfg *config.Logger) logger.Option {
	switch strings.ToLower(cfg.Level) {
	case "debug":
		return logger.WithLevel(logger.LevelDebug)
	case "info":
		return logger.WithLevel(logger.LevelInfo)
	case "warn":
		return logger.WithLevel(logger.LevelWarn)
	case "error":
		return logger.WithLevel(logger.LevelError)
	case "panic":
		return logger.WithLevel(logger.LevelPanic)
	case "fatal":
		return logger.WithLevel(logger.LevelFatal)
	default:
		return nil
	}
}

func formatOption(cfg *config.Logger) logger.Option {
	switch strings.ToLower(cfg.Format) {
	case "json":
		return logger.WithFormat(logger.FormatJSON)
	case "text":
		return logger.WithFormat(logger.FormatText)
	default:
		return nil
	}
}

func timeLayoutOption(cfg *config.Logger) logger.Option {
	switch strings.ToLower(cfg.TimeLayout) {
	case "datetime":
		return logger.WithTimeLayout(logger.TimeLayoutDateTime)
	case "date":
		return logger.WithTimeLayout(logger.TimeLayoutDateOnly)
	case "time":
		return logger.WithTimeLayout(logger.TimeLayoutTimeOnly)
	case "rfc3339":
		return logger.WithTimeLayout(logger.TimeLayoutRFC3339)
	case "rfc822":
		return logger.WithTimeLayout(logger.TimeLayoutRFC822)
	case "rfc1123":
		return logger.WithTimeLayout(logger.TimeLayoutRFC1123)
	default:
		return nil
	}
}

func fileRotationOption(cfg *config.Logger) logger.Option {
	if cfg.FileRotation.MaxAge > 0 && cfg.FileRotation.MaxSize > 0 && cfg.FileRotation.MaxBackups > 0 {
		return logger.WithFileRotation(
			cfg.FileRotation.MaxAge,
			cfg.FileRotation.MaxSize,
			cfg.FileRotation.MaxBackups,
			cfg.FileRotation.LocalTime,
			cfg.FileRotation.Compress,
		)
	}

	return nil
}
