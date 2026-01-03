package zap

import (
	"errors"
	"go-auth/pkg/logger"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	logger.Register(logger.DriverZap, New)
}

type adapter struct {
	l *zap.SugaredLogger
}

func New(cfg logger.Config) (logger.Logger, error) {
	if err := ensureDirectories(cfg.OutputPaths); err != nil {
		return nil, err
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.DateTime),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.Development && cfg.Encoding == logger.EncodingConsole {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.Level(cfg.Level)),
		Development:       cfg.Development,
		DisableCaller:     cfg.DisableCaller,
		DisableStacktrace: cfg.DisableStacktrace,
		EncoderConfig:     encoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	if cfg.Encoding == logger.EncodingConsole {
		zapCfg.Encoding = string(logger.EncodingConsole)
	} else {
		zapCfg.Encoding = string(logger.EncodingJSON)
	}

	if len(cfg.OutputPaths) > 0 {
		zapCfg.OutputPaths = cfg.OutputPaths
	}

	if len(cfg.InitialFields) > 0 {
		zapCfg.InitialFields = cfg.InitialFields
	}

	if cfg.DriverOptions != nil {
		if val, ok := cfg.DriverOptions[samplingKey]; ok {
			if samplingOpts, ok := val.(SamplingConfig); ok {
				zapCfg.Sampling = &zap.SamplingConfig{
					Initial:    samplingOpts.Initial,
					Thereafter: samplingOpts.Thereafter,
				}
			}
		}
	}

	z, err := zapCfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &adapter{l: z.Sugar()}, nil

}

func (a *adapter) Info(msg string, args ...any)   { a.l.Infow(msg, args...) }
func (a *adapter) Debug(msg string, args ...any)  { a.l.Debugw(msg, args...) }
func (a *adapter) Warn(msg string, args ...any)   { a.l.Warnw(msg, args...) }
func (a *adapter) Error(msg string, args ...any)  { a.l.Errorw(msg, args...) }
func (a *adapter) Fatal(msg string, args ...any)  { a.l.Fatalw(msg, args...) }
func (a *adapter) With(args ...any) logger.Logger { return &adapter{l: a.l.With(args...)} }
func (a *adapter) Sync() error {
	err := a.l.Sync()
	if err != nil && !isIgnorableSyncErr(err) {
		return err
	}
	return nil
}

func ensureDirectories(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	for _, path := range paths {
		if path == "stdout" || path == "stderr" {
			continue
		}

		dir := filepath.Dir(path)
		if dir != "." && dir != "/" {
			if err := os.MkdirAll(dir, 0o750); err != nil {
				return err
			}
		}
	}
	return nil
}

// isIgnorableSyncErr: Linux/Mac'te stdout sync hatalarını yutmak için.
func isIgnorableSyncErr(err error) bool {
	if err == nil {
		return false
	}

	// Doğrudan syscall hatası mı?
	var errno syscall.Errno
	if errors.As(err, &errno) && isIgnorableErrno(errno) {
		return true
	}

	// PathError içine sarılmış mı?
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		if errors.As(pathErr.Err, &errno) && isIgnorableErrno(errno) {
			return true
		}
	}
	return false
}

func isIgnorableErrno(errno syscall.Errno) bool {
	switch errno {
	// ENOTTY: Inappropriate ioctl for device (genelde stdout redirect edildiğinde çıkar)
	case syscall.ENOTTY, syscall.EINVAL, syscall.EPIPE, syscall.ENOSYS:
		return true
	default:
		return false
	}
}
