package main

import (
	"context"
	"fmt"
	"os"

	"go-auth/internal/bootstrap"
	"go-auth/internal/config"
	"go-auth/internal/handler"
	"go-auth/internal/repository"
	"go-auth/internal/security"
	"go-auth/internal/service"
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

	userRepo, sessionRepo, _ := repository.NewRepositories(pool)
	passwordHasher := security.NewHasher(cfg.Security.HashCost)
	opaqueTokenManager := security.NewOpaque(32)

	accessTokenManager, err := security.NewJWT(cfg.Security.JWTSecret, cfg.App.Name, cfg.Security.AccessTTL)
	if err != nil {
		return fmt.Errorf("create access token manager: %w", err)
	}

	svc, err := service.NewService(&service.Config{
		UserRepo:           userRepo,
		SessionRepo:        sessionRepo,
		PasswordHasher:     passwordHasher,
		OpaqueTokenManager: opaqueTokenManager,
		AccessTokenManager: accessTokenManager,
		AccessTokenTTL:     cfg.Security.AccessTTL,
		RefreshTokenTTL:    cfg.Security.RefreshTTL,
	})
	if err != nil {
		return fmt.Errorf("create service: %w", err)
	}

	authHandler := handler.NewAuthHandler(svc)
	healthHandler := handler.NewHealthHandler()
	router := handler.NewRouter(ctx, handler.Deps{
		Auth:   authHandler,
		Cfg:    cfg,
		Health: healthHandler,
		Log:    log,
	})

	if err := bootstrap.RunServer(cfg, router, log); err != nil {
		return fmt.Errorf("server: %w", err)
	}

	return nil
}
