// Package main is the main package for the app.
package main

import (
	"os"
	"os/signal"
	"syscall"

	"go-auth/pkg/logger"
	"go-auth/pkg/logger/zap"
)

func main() {
	zapLogger, err := zap.New(zap.WithDevelopmentMode())
	if err != nil {
		panic(err)
	}

	logger.SetGlobalLogger(zapLogger)

	defer func() {
		_ = logger.Sync()
	}()

	logger.Debug("Application debug message")
	logger.Info("Application info message")
	logger.Warn("Application warn message")
	logger.With("key", "value").Debug("Application debug message with key-value")
	logger.With("key", "value").With("key2", "value2").Debug("Application debug message with key-value and key2-value2")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")
}
