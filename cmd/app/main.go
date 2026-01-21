package main

import (
	"fmt"
	"os"

	"go-auth/internal/bootstrap"
	"go-auth/internal/config"
	"go-auth/pkg/logger"
	_ "go-auth/pkg/logger/adapter/zap"
	_ "go-auth/pkg/logger/adapter/zerolog"
)

func main() {
	cfg, err := config.NewLoader().Load(".env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(bootstrap.BuildLoggerOptions(cfg)...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	defer func() { _ = log.Sync() }()

	log.Info("Application started",
		"app_name", cfg.App.Name,
		"env", cfg.App.Env,
		"server_address", cfg.ServerAddr(),
		"database_url", cfg.DatabaseURL(),
		"smtp_address", cfg.SMTPAddr(),
		"jwt_key_length", len(cfg.JWTKey()),
	)
}
