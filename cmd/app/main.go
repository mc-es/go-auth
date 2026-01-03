// Package main is the main package for the app.
package main

import (
	"os"
	"os/signal"
	"syscall"

	"go-auth/internal/config"
	"go-auth/pkg/logger"
	"go-auth/pkg/logger/logrus"
	"go-auth/pkg/logger/zap"
)

func main() {
	_, err := config.LoadEnv()
	if err != nil {
		panic(err)
	}

	zap.Register()
	logrus.Register()

	log, err := logger.New(
		logger.WithDriver(logger.DriverLogrus),
		logger.WithFormatter(logger.FormatterText),
		logger.WithTimeLayout(logger.TimeLayoutRFC3339),
		logger.WithDevMode(),
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = log.Sync()
	}()

	log.Debug("Debug message - should be visible", "app", "auth", "driver", "zap")
	log.Info("Info message - should be visible", "app", "auth", "driver", "zap")
	log.Warn("Warn message - should be visible", "app", "auth", "driver", "zap")
	log.Error("Error message - should be visible", "app", "auth", "driver", "zap")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}
