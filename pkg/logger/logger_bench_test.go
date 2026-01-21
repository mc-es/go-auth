package logger_test

import (
	"context"
	"os"
	"testing"

	"go-auth/pkg/logger"
	_ "go-auth/pkg/logger/adapter/zap"
	_ "go-auth/pkg/logger/adapter/zerolog"
)

func BenchmarkZapLogger(b *testing.B) {
	benchmarkLogger(b, logger.DriverZap)
}

func BenchmarkZerologLogger(b *testing.B) {
	benchmarkLogger(b, logger.DriverZerolog)
}

func benchmarkLogger(b *testing.B, driver logger.Driver) {
	logFile := createTempFile(b)
	ctx := context.Background()

	// Setup Logger with Info level
	logInfo, _ := logger.New(
		logger.WithDriver(driver),
		logger.WithLevel(logger.LevelInfo),
		logger.WithFormat(logger.FormatJSON),
		logger.WithOutputPaths(logFile),
	)

	// Setup Logger with Debug level
	logDebug, _ := logger.New(
		logger.WithDriver(driver),
		logger.WithLevel(logger.LevelDebug),
		logger.WithFormat(logger.FormatJSON),
		logger.WithOutputPaths(logFile),
	)

	b.Run("log info", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logInfo.Info("benchmark")
		}
	})

	b.Run("log info with fields", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logInfo.Info("benchmark", "key1", "value1", "key2", "value2")
		}
	})

	b.Run("log info with context", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logInfo.InfoCtx(ctx, "benchmark", "key1", "value1")
		}
	})

	b.Run("log info with child logger", func(b *testing.B) {
		child := logInfo.Named("sub-component")

		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			child.Info("benchmark", "key", "value")
		}
	})

	b.Run("log debug", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			logDebug.Debug("benchmark", "key1", "value1")
		}
	})
}

func createTempFile(b *testing.B) string {
	b.Helper()

	file, err := os.CreateTemp("", "bench-*.log")
	if err != nil {
		b.Fatal(err)
	}

	_ = file.Close()
	name := file.Name()

	b.Cleanup(func() { _ = os.Remove(name) })

	return name
}
