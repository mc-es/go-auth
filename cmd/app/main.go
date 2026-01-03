package main

import (
	"go-auth/pkg/logger"

	"go-auth/pkg/logger/zap"
)

func main() {
	log, err := logger.New(
		logger.WithDriver(logger.DriverZap),
		logger.WithLevel(logger.LevelDebug),
		logger.WithEncoding(logger.EncodingConsole),
		logger.WithDevelopment(),
		zap.WithSampling(100, 1000),
	)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	log.Info("Uygulama ayağa kalktı", "port", 8080)

	log.With("request_id", "abc-123", "ip", "127.0.0.1").Debug("Veritabanı bağlantısı kuruluyor...")
}
