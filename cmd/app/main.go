// Package main is the main package for the app.
package main

import (
	"os"
	"os/signal"
	"syscall"

	"go-auth/internal/config"
	"go-auth/pkg/logger"
	"go-auth/pkg/logger/logrus"
)

func main() {
	logrusLogger, err := logrus.New(logrus.WithDevelopmentMode())
	if err != nil {
		panic(err)
	}

	logger.SetGlobalLogger(logrusLogger)

	defer func() {
		_ = logger.Sync()
	}()

	envConfig, err := config.LoadEnv()
	if err != nil {
		logger.Fatal("Failed to load environment configuration", "error", err)
	}

	logger.Warn(
		"Application started",
		"name",
		envConfig.AppName,
		"host",
		envConfig.Host,
		"port",
		envConfig.Port,
		"env",
		envConfig.Env,
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")
}
