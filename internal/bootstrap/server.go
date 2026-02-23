package bootstrap

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go-auth/internal/config"
	"go-auth/pkg/logger"
)

func RunServer(cfg *config.Config, handler http.Handler, log logger.Logger) error {
	srv := &http.Server{
		Addr:         cfg.ServerAddr(),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTO,
		WriteTimeout: cfg.Server.WriteTO,
		IdleTimeout:  cfg.Server.IdleTO,
	}

	go func() {
		log.Info("HTTP server listening", "addr", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTO)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	log.Info("HTTP server stopped")

	return nil
}
