// Package main is the main package for the app.
package main

import (
	"os"
	"os/signal"
	"syscall"

	"go-auth/internal/config"
)

func main() {
	_, err := config.LoadEnv()
	if err != nil {
		panic(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}
