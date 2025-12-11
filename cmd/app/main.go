// Package app is the app's entry point and starts auth service.
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
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

		_ = logger.Sync()

		os.Exit(1)
	}

	logger.S().Infow("Application stopped gracefully")

	_ = logger.Sync()
}

func initLogger(cfg *config.Config) error {
	isDev := cfg.Server.Env == "dev"

	opts := []logger.Option{
		logger.WithInitialFields(map[string]any{
			"app_env":     cfg.Server.Env,
			"app_name":    cfg.Server.AppName,
			"app_version": cfg.Server.Version,
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

	db, err := database.NewDatabase(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("database init: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.S().Warnw("Database connection close failed during cleanup", "error", err)
		}
	}()

	monitor, err := database.NewMonitor(
		db.Ping,
		cfg.Database.Resilience.HealthCheckInterval,
		cfg.Database.Resilience.HealthCheckTimeout,
	)
	if err != nil {
		return fmt.Errorf("monitor init: %w", err)
	}

	monitor.Start(ctx)

	server := newServer(cfg.Server)

	logger.S().Infow("Server starting...",
		"address", server.Addr,
	)

	serverErrors := make(chan error, 1)
	go startServerListener(server, serverErrors)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	return waitForShutdown(
		cancel,
		server,
		monitor,
		db,
		serverErrors,
		quit,
		cfg.Server.ShutdownTimeout,
	)
}

func newServer(cfg config.Server) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)

	return &http.Server{
		Addr:         net.JoinHostPort(cfg.Host, strconv.FormatUint(cfg.Port, 10)),
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}

func rootHandler(res http.ResponseWriter, _ *http.Request) {
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
	monitor database.Monitor,
	db database.Database,
	serverErrors <-chan error,
	quit <-chan os.Signal,
	shutdownTO time.Duration,
) error {
	select {
	case err, ok := <-serverErrors:
		if !ok {
			cancel()

			return nil
		}

		logger.S().Errorw("Server failed", "error", err)

		monitor.Stop()

		cancel()

		if err := db.Close(); err != nil {
			logger.S().Warnw("Database shutdown failed", "error", err)
		}

		return err
	case sig := <-quit:
		return shutdownServer(cancel, server, monitor, db, serverErrors, sig, shutdownTO)
	}
}

func shutdownServer(
	cancel context.CancelFunc,
	server *http.Server,
	monitor database.Monitor,
	db database.Database,
	serverErrors <-chan error,
	sig os.Signal,
	shutdownTO time.Duration,
) error {
	logger.S().Infow("Shutdown signal received", "signal", sig.String())

	serverCtx, serverCancel := context.WithTimeout(context.Background(), shutdownTO)
	defer serverCancel()

	var shutdownErr error

	if err := server.Shutdown(serverCtx); err == nil {
		logger.S().Infow("Server traffic stopped")
	} else if errors.Is(err, context.DeadlineExceeded) {
		logger.S().Warnw("Server shutdown timed out, forcing exit", "timeout", shutdownTO)

		shutdownErr = err
	} else {
		logger.S().Errorw("Server shutdown failed", "error", err)
		shutdownErr = err
	}

	cancel()
	monitor.Stop()

	if err := db.Close(); err != nil {
		logger.S().Warnw("Database shutdown failed", "error", err)

		if shutdownErr == nil {
			shutdownErr = err
		}
	} else {
		logger.S().Infow("Database connection closed")
	}

	select {
	case err := <-serverErrors:
		if err != nil {
			logger.S().Errorw("Server error during shutdown", "error", err)

			if shutdownErr == nil {
				shutdownErr = err
			}
		}
	default:
	}

	return shutdownErr
}
