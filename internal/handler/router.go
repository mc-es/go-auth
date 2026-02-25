package handler

import (
	"context"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"go-auth/internal/config"
	"go-auth/internal/middleware"
	"go-auth/pkg/logger"
)

type Deps struct {
	Auth   *AuthHandler
	Cfg    *config.Config
	Health *HealthHandler
	Log    logger.Logger
}

func NewRouter(ctx context.Context, deps Deps) chi.Router {
	r := chi.NewRouter()

	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(middleware.Logger(deps.Log))
	r.Use(middleware.CORS(&deps.Cfg.CORS))
	r.Use(middleware.RateLimit(ctx, deps.Cfg.RateLimit))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", deps.Auth.Register)
		r.Post("/login", deps.Auth.Login)
		r.Post("/logout", deps.Auth.Logout)
		r.Post("/refresh", deps.Auth.Refresh)
	})

	r.Get("/health", deps.Health.Liveness)
	r.Get("/health/live", deps.Health.Liveness)

	return r
}
