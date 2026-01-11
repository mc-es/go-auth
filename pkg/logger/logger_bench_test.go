package logger_test

import (
	"testing"

	"go-auth/pkg/logger"
	_ "go-auth/pkg/logger/adapter/logrus"
	_ "go-auth/pkg/logger/adapter/zap"
)

func BenchmarkZapLogger(b *testing.B) {
	benchmarkLogger(b, "zap", logger.DriverZap)
}

func BenchmarkLogrusLogger(b *testing.B) {
	benchmarkLogger(b, "logrus", logger.DriverLogrus)
}

func benchmarkLogger(b *testing.B, name string, driver logger.Driver) {
	// Setup Logger
	log, _ := logger.New(
		logger.WithDriver(driver),
		logger.WithLevel(logger.LevelInfo),
		logger.WithFormat(logger.FormatJSON),
		logger.WithOutputPaths("/dev/null"), // Discard output
		logger.WithFileRotation(1, 1000000, 1, false, false),
	)

	b.Run(name+" clean strings", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			log.Info("benchmark", "key1", "value1", "key2", "value2")
		}
	})

	b.Run(name+" dirty optimization miss", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Odd number of args triggers dirty flag -> allocation happens
			log.Info("benchmark", "key1", "value1", "key2")
		}
	})

	b.Run(name+" non string key", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Non-string key triggers dirty flag -> allocation happens
			log.Info("benchmark", "key1", "value1", 123, "val")
		}
	})
}
