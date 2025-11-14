// Package main is the app's entry point and starts auth service.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/mc-es/go-auth/pkg/logger"
)

const defaultPort = 8080
const (
	readTimeout     = 5 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 120 * time.Second
	shutdownTimeout = 10 * time.Second
)

func main() {
	if err := logger.Init(
		logger.WithDevelopmentMode(),
		logger.WithLevel("debug"),
		logger.WithInitialFields(map[string]any{
			"app":     "go-auth",
			"version": "1.0.0",
		}),
	); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "logger init failed: %v\n", err)

		os.Exit(1)
	}

	if err := godotenv.Load(); err != nil {
		logger.S().Warnw("No .env file found, continuing with existing environment")
	}

	if err := run(); err != nil {
		logger.S().Errorw("Application stopped with error", "error", err)

		_ = logger.Sync()

		os.Exit(1)
	}

	_ = logger.Sync()
}

func getPort() int {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil || port <= 0 {
		logger.S().Warnw("Invalid PORT value, using default", "PORT", os.Getenv("PORT"))

		return defaultPort
	}

	return port
}

func run() error {
	port := getPort()
	server := newServer(port)

	logger.S().Infow("Server starting...", "port", port)

	serverErrors := make(chan error, 1)
	go startServerListener(server, serverErrors)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	return waitForShutdown(server, serverErrors, quit)
}

func newServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	logger.S().Infow("Incoming request",
		"method", req.Method,
		"path", req.URL.Path,
		"remote_addr", req.RemoteAddr,
	)

	_, _ = fmt.Fprintf(res, "Hello from go-auth 👋")
}

func startServerListener(server *http.Server, serverErrors chan<- error) {
	defer close(serverErrors)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		serverErrors <- err
	}
}

func waitForShutdown(server *http.Server, serverErrors <-chan error, quit <-chan os.Signal) error {
	select {
	case err, ok := <-serverErrors:
		if ok && err != nil {
			logger.S().Errorw("Server failed", "error", err)

			return err
		}

		logger.S().Infow("Server stopped without signal")

		return nil

	case sig := <-quit:
		return shutdownServer(server, serverErrors, sig)
	}
}

func shutdownServer(server *http.Server, serverErrors <-chan error, sig os.Signal) error {
	logger.S().Infow("Shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.S().Errorw(
				"Forced shutdown: context deadline exceeded",
				"timeout", shutdownTimeout,
			)
		} else {
			logger.S().Errorw("Graceful shutdown failed", "error", err)
		}

		return err
	}

	logger.S().Infow("Server stopped gracefully")

	if err, ok := <-serverErrors; ok && err != nil {
		logger.S().Errorw("Server failed during shutdown", "error", err)

		return err
	}

	return nil
}
