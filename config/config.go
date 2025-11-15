// Package config defines application configuration structures and helpers.
package config

import (
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

// ServerConfig contains runtime information for the HTTP server.
type ServerConfig struct {
	App  string `env:"SERVER_APP"  validate:"required"`
	Host string `env:"SERVER_HOST" validate:"hostname|ip"    envDefault:"0.0.0.0"`
	Port uint   `env:"SERVER_PORT" validate:"port"           envDefault:"8080"`
	Mode string `env:"SERVER_MODE" validate:"oneof=dev prod" envDefault:"dev"`
}

// DatabaseConfig stores connection parameters for the primary database.
type DatabaseConfig struct {
	URL            string        `env:"DB_URL"             validate:"required,mongodb_connection_string"`
	MaxConnections int           `env:"DB_MAX_CONNECTIONS" validate:"min=0,max=30"                       envDefault:"10"`
	Timeout        time.Duration `env:"DB_TIMEOUT"         validate:"min=0,max=30s"                      envDefault:"5s"`
}

// AuthConfig configures access and refresh token lifetimes.
type AuthConfig struct {
	AccessTTL  time.Duration `env:"AUTH_ACCESS_TTL"  validate:"min=0,max=1h"    envDefault:"15m"`
	RefreshTTL time.Duration `env:"AUTH_REFRESH_TTL" validate:"min=0,max=1000h" envDefault:"720h"` // 30 days
}

// SecurityConfig holds security-related secrets and parameters.
type SecurityConfig struct {
	JWTSecret string `env:"SECURITY_JWT_SECRET" validate:"required,min=24"`
	HashCost  int    `env:"SECURITY_HASH_COST"  validate:"min=8,max=16"    envDefault:"12"`
}

// RateLimitConfig limits repeated requests from a source.
type RateLimitConfig struct {
	Requests int           `env:"RATE_LIMIT_REQUESTS" validate:"min=0,max=10" envDefault:"5"`
	Duration time.Duration `env:"RATE_LIMIT_DURATION" validate:"min=0,max=2m" envDefault:"1m"`
}

// SMTPConfig contains credentials and settings for outgoing email.
type SMTPConfig struct {
	Host     string `env:"SMTP_HOST"     validate:"required,hostname|ip"`
	Port     uint   `env:"SMTP_PORT"     validate:"required,port"`
	Username string `env:"SMTP_USERNAME" validate:"required"`
	Password string `env:"SMTP_PASSWORD" validate:"required"`
	From     string `env:"SMTP_FROM"     validate:"required,email"`
}

// CORSConfig defines cross-origin access rules for the API.
type CORSConfig struct {
	AllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," validate:"dive"`
	AllowedMethods []string `env:"CORS_ALLOWED_METHODS" envSeparator:"," validate:"dive,oneof=GET POST PUT DELETE OPTIONS"`
}

// Config bundles all configuration sections loaded from the environment.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Security SecurityConfig
	Rate     RateLimitConfig
	SMTP     SMTPConfig
	CORS     CORSConfig
}

var (
	validate     *validator.Validate
	validateOnce sync.Once
)

// Load reads environment variables into a Config instance.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	setupValidator()

	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// setupValidator initializes the validator instance and attaches struct-level validations.
func setupValidator() {
	validateOnce.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
		validate.RegisterStructValidation(authConfigStructLevelValidation, AuthConfig{})
	})
}

// authConfigStructLevelValidation enforces that refresh tokens live longer than access tokens.
func authConfigStructLevelValidation(sl validator.StructLevel) {
	ac := sl.Current().Interface().(AuthConfig)

	if ac.RefreshTTL <= ac.AccessTTL {
		sl.ReportError(ac.RefreshTTL, "RefreshTTL", "RefreshTTL", "gtfield", "AccessTTL")
	}
}
