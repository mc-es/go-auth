package main

import (
	"context"
	"fmt"
	"os"

	"go-auth/internal/bootstrap"
	"go-auth/internal/config"
	"go-auth/internal/repository"
	"go-auth/pkg/logger"
	_ "go-auth/pkg/logger/adapter/zap"
	_ "go-auth/pkg/logger/adapter/zerolog"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.NewLoader().Load(".env")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log, err := logger.New(bootstrap.BuildLoggerOptions(cfg)...)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}

	defer func() { _ = log.Sync() }()

	ctx := context.Background()

	pool, err := bootstrap.NewDBPool(ctx, cfg)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	userRepo, sessionRepo := repository.NewRepositories(pool)

	_ = userRepo
	_ = sessionRepo

	log.Info("Repository layer ready",
		"user_repo", "postgres",
		"session_repo", "postgres",
	)

	return nil
}
