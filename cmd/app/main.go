// Package main is the main package for the app.
package main

import (
	"os"
	"os/signal"
	"syscall"

	"go-auth/internal/config"
	"go-auth/pkg/logger"
	"go-auth/pkg/logger/zap"
)

func main() {
	zapLogger, err := zap.New(zap.WithDevelopmentMode(), zap.WithoutStacktrace())
	if err != nil {
		panic(err)
	}

	logger.SetGlobalLogger(zapLogger)

	defer func() {
		_ = logger.Sync()
	}()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
	}

	logger.Info("Application started", "name", cfg.AppName, "host", cfg.Host, "port", cfg.Port, "env", cfg.Env)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")
}
