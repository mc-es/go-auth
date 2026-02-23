package handler

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"go-auth/internal/config"
	"go-auth/internal/middleware"
	"go-auth/pkg/logger"
)

func NewRouter(auth *AuthHandler, health *HealthHandler, cfg *config.Config, log logger.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(middleware.Logger(log))
	r.Use(middleware.CORS(&cfg.CORS))
	r.Use(middleware.RateLimit(cfg.RateLimit))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", auth.Register)
		r.Post("/login", auth.Login)
		r.Post("/logout", auth.Logout)
		r.Post("/refresh", auth.Refresh)
	})

	r.Get("/health", health.Liveness)
	r.Get("/health/live", health.Liveness)

	return r
}
