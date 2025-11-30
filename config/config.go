// Package config defines application configuration structures and helpers.
package config

import (
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

// Server holds runtime information for the HTTP server.
type Server struct {
	App     string `env:"SERVER_APP"     validate:"required"`
	Version string `env:"SERVER_VERSION" validate:"required"`
	Host    string `env:"SERVER_HOST"    validate:"hostname|ip"    envDefault:"0.0.0.0"`
	Port    uint   `env:"SERVER_PORT"    validate:"port"           envDefault:"8080"`
	Mode    string `env:"SERVER_MODE"    validate:"oneof=dev prod" envDefault:"dev"`
}

// Database stores connection parameters for the primary database.
type Database struct {
	Name          string        `env:"DB_NAME"            validate:"required"`
	URL           string        `env:"DB_URL"             validate:"required,mongodb_connection_string"`
	MaxConns      uint64        `env:"DB_MAX_CONNS"       validate:"min=1,max=200,gtfield=MinConns"     envDefault:"100"`
	MinConns      uint64        `env:"DB_MIN_CONNS"       validate:"min=0,max=50"                       envDefault:"20"`
	MaxIdleTime   time.Duration `env:"DB_MAX_IDLE_TIME"   validate:"min=0,max=30m"                      envDefault:"30m"`
	ConnectTO     time.Duration `env:"DB_CONNECT_TO"      validate:"min=0,max=30s"                      envDefault:"10s"`
	SelectionTO   time.Duration `env:"DB_SELECTION_TO"    validate:"min=0,max=30s"                      envDefault:"5s"`
	PingTO        time.Duration `env:"DB_PING_TO"         validate:"min=0,max=30s"                      envDefault:"5s"`
	CloseTO       time.Duration `env:"DB_CLOSE_TO"        validate:"min=0,max=60s"                      envDefault:"10s"`
	HealthCheckIT time.Duration `env:"DB_HEALTH_CHECK_IT" validate:"min=1m,max=60m"                     envDefault:"10m"`
	HealthCheckTO time.Duration `env:"DB_HEALTH_CHECK_TO" validate:"min=1s,max=30s,gtfield=PingTO"      envDefault:"10s"`
}

// Auth stores access and refresh token lifetimes.
type Auth struct {
	AccessTTL  time.Duration `env:"AUTH_ACCESS_TTL"  validate:"min=0,max=1h,ltfield=RefreshTTL"   envDefault:"15m"`
	RefreshTTL time.Duration `env:"AUTH_REFRESH_TTL" validate:"min=0,max=1000h,gtfield=AccessTTL" envDefault:"720h"`
}

// Security holds security-related secrets and parameters.
type Security struct {
	JWTSecret string `env:"SECURITY_JWT_SECRET" validate:"required,min=24"`
	HashCost  int    `env:"SECURITY_HASH_COST"  validate:"min=8,max=16"    envDefault:"12"`
}

// RateLimit limits repeated requests from a source.
type RateLimit struct {
	Requests int           `env:"RATE_LIMIT_REQUESTS" validate:"min=0,max=10" envDefault:"5"`
	Duration time.Duration `env:"RATE_LIMIT_DURATION" validate:"min=0,max=2m" envDefault:"1m"`
}

// SMTP contains credentials and settings for outgoing email.
type SMTP struct {
	Host     string `env:"SMTP_HOST"     validate:"required,hostname|ip"`
	Port     uint   `env:"SMTP_PORT"     validate:"required,port"`
	Username string `env:"SMTP_USERNAME" validate:"required"`
	Password string `env:"SMTP_PASSWORD" validate:"required"`
	From     string `env:"SMTP_FROM"     validate:"required,email"`
}

// CORS defines cross-origin access rules for the API.
type CORS struct {
	AllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," validate:"dive"`
	AllowedMethods []string `env:"CORS_ALLOWED_METHODS" envSeparator:"," validate:"dive,oneof=GET POST PUT DELETE OPTIONS"`
}

// Config stores all configuration sections loaded from the environment.
type Config struct {
	Server   Server
	Database Database
	Auth     Auth
	Security Security
	Rate     RateLimit
	SMTP     SMTP
	CORS     CORS
}

var (
	validate     *validator.Validate
	validateOnce sync.Once
)

// Load reads environment variables into a Config struct.
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

// setupValidator initializes the validator struct.
func setupValidator() {
	validateOnce.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
	})
}
