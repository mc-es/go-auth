package main

import (
	"go-auth/pkg/logger"
	_ "go-auth/pkg/logger/adapter/logrus"
	_ "go-auth/pkg/logger/adapter/zap"
)

func main() {
	log, err := logger.New(
		logger.WithDriver(logger.DriverLogrus),
		logger.WithDevelopment(),
	)
	if err != nil {
		panic(err)
	}

	defer func() { _ = log.Sync() }()

	log.Info("Hello, World!", logger.A("key", "value"))
}
