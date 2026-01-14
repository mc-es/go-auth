package nop_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-auth/pkg/logger"
	_ "go-auth/pkg/logger/adapter/nop"
)

func TestNopLogger(t *testing.T) {
	log, err := logger.New(logger.WithDriver(logger.DriverNop))
	require.NoError(t, err)
	require.NotNil(t, log)

	ctx := context.Background()

	assert.NotPanics(t, func() {
		log.Debug("debug message", "key", "value")
		log.Info("info message", "key", "value")
		log.Warn("warn message", "key", "value")
		log.Error("error message", "key", "value")
		log.Panic("panic message", "key", "value") // Should NOT panic
		log.Fatal("fatal message", "key", "value") // Should NOT exit

		log.DebugCtx(ctx, "debug message", "key", "value")
		log.InfoCtx(ctx, "info message", "key", "value")
		log.WarnCtx(ctx, "warn message", "key", "value")
		log.ErrorCtx(ctx, "error message", "key", "value")
		log.PanicCtx(ctx, "panic message", "key", "value") // Should NOT panic
		log.FatalCtx(ctx, "fatal message", "key", "value") // Should NOT exit
	})

	// Sync should always return nil
	err = log.Sync()
	assert.NoError(t, err)
}

func TestNopLoggerIgnoresConfig(t *testing.T) {
	// NOP logger should work with any configuration
	log, err := logger.New(
		logger.WithDriver(logger.DriverNop),
		logger.WithLevel(logger.LevelError),
		logger.WithFormat(logger.FormatJSON),
		logger.WithOutputPaths("stdout", "stderr"),
		logger.WithFileRotation(1, 100, 5, true, true),
	)
	require.NoError(t, err)
	require.NotNil(t, log)

	// Should not panic regardless of config
	assert.NotPanics(t, func() {
		log.Info("test message")
		log.Error("test error")
	})
}
