// Package app is the app's entry point and starts auth service.
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

	"go-auth/config"
	"go-auth/pkg/database"
	"go-auth/pkg/logger"
)

const (
	readTimeout     = 5 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 120 * time.Second
	shutdownTimeout = 10 * time.Second
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "config load failed: %v\n", err)

		os.Exit(1)
	}

	if err := initLogger(cfg); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "logger init failed: %v\n", err)

		os.Exit(1)
	}

	err = run(cfg)
	if err != nil {
		logger.S().Errorw("Application stopped with error", "error", err)
	}

	_ = logger.Sync()

	if err != nil {
		os.Exit(1)
	}
}

func initLogger(cfg *config.Config) error {
	isDev := cfg.Server.Mode == "dev"

	opts := []logger.Option{
		logger.WithInitialFields(map[string]any{
			"app":     cfg.Server.App,
			"version": "1.0.0",
			"mode":    cfg.Server.Mode,
		}),
	}

	if isDev {
		opts = append(opts,
			logger.WithDevelopmentMode(),
			logger.WithLevel(logger.LevelDebug),
		)
	} else {
		opts = append(opts,
			logger.WithLevel(logger.LevelInfo),
		)
	}

	return logger.Init(opts...)
}

func run(cfg *config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbInstance, err := database.NewDatabase(cfg.Database)
	if err != nil {
		logger.S().Errorw("Database initialization failed", "error", err)

		return err
	}

	dbInstance.StartHealthCheck(ctx)

	server := newServer(cfg.Server)

	logger.S().Infow("Server starting...",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
	)

	serverErrors := make(chan error, 1)
	go startServerListener(server, serverErrors)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	return waitForShutdown(cancel, server, dbInstance, serverErrors, quit)
}

func newServer(cfg config.ServerConfig) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)

	return &http.Server{
		Addr:         cfg.Host + ":" + strconv.FormatUint(uint64(cfg.Port), 10),
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

func waitForShutdown(
	cancel context.CancelFunc,
	server *http.Server,
	dbInstance *database.Database,
	serverErrors <-chan error,
	quit <-chan os.Signal,
) error {
	select {
	case err, ok := <-serverErrors:
		if ok && err != nil {
			logger.S().Errorw("Server failed", "error", err)
			cancel()

			return err
		}

		logger.S().Infow("Server stopped without signal")
		cancel()

		return nil

	case sig := <-quit:
		return shutdownServer(cancel, server, dbInstance, serverErrors, sig)
	}
}

func shutdownServer(
	cancel context.CancelFunc,
	server *http.Server,
	dbInstance *database.Database,
	serverErrors <-chan error,
	sig os.Signal,
) error {
	logger.S().Infow("Shutdown signal received", "signal", sig.String())

	cancel()

	dbCtx, dbCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer dbCancel()

	if err := dbInstance.Close(dbCtx); err != nil {
		logger.S().Errorw("Database shutdown failed", "error", err)
	}

	serverCtx, serverCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer serverCancel()

	if err := server.Shutdown(serverCtx); err != nil {
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
