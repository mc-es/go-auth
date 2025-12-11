// Package config defines application configuration structures and loaders.
package config

import (
	"sync"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var (
	validate     *validator.Validate
	validateOnce sync.Once
)

// Load reads environment variables into a Config struct and validates them.
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

// setupValidator initializes the validator instance.
func setupValidator() {
	validateOnce.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
	})
}

// Config stores all configuration sections loaded from the environment.
type Config struct {
	Server   Server    `envPrefix:"SERVER_"`
	Database Database  `envPrefix:"DB_"`
	Auth     Auth      `envPrefix:"AUTH_"`
	Security Security  `envPrefix:"SECURITY_"`
	Rate     RateLimit `envPrefix:"RATE_LIMIT_"`
	SMTP     SMTP      `envPrefix:"SMTP_"`
	CORS     CORS      `envPrefix:"CORS_"`
}

// Server holds runtime information for the HTTP server.
type Server struct {
	AppName         string        `env:"APP_NAME"         validate:"required"`
	Version         string        `env:"VERSION"          validate:"required"`
	Host            string        `env:"HOST"             validate:"hostname|ip"    envDefault:"0.0.0.0"`
	Port            uint64        `env:"PORT"             validate:"port"           envDefault:"8080"`
	Env             string        `env:"ENV"              validate:"oneof=dev prod" envDefault:"dev"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT"     validate:"min=0,max=30s"  envDefault:"5s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT"    validate:"min=0,max=30s"  envDefault:"10s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT"     validate:"min=0,max=5m"   envDefault:"120s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" validate:"min=0,max=30s"  envDefault:"10s"`
}

// Database holds all database configuration sections.
type Database struct {
	Connection dbConnection `envPrefix:"CONNECTION_"`
	Pool       dbPool       `envPrefix:"POOL_"`
	Resilience dbResilience `envPrefix:"RESILIENCE_"`
	Timeout    dbTimeout    `envPrefix:"TIMEOUT_"`
}

// Auth stores access and refresh token lifetimes.
type Auth struct {
	AccessTTL  time.Duration `env:"ACCESS_TTL"  validate:"min=0,max=1h,ltfield=RefreshTTL" envDefault:"15m"`
	RefreshTTL time.Duration `env:"REFRESH_TTL" validate:"min=0,max=1000h"                 envDefault:"720h"`
}

// Security holds security-related secrets and parameters.
type Security struct {
	JWTSecret string `env:"JWT_SECRET" validate:"required,min=24"`
	HashCost  int    `env:"HASH_COST"  validate:"min=8,max=16"    envDefault:"12"`
}

// RateLimit limits repeated requests from a source.
type RateLimit struct {
	MaxRequests int           `env:"MAX_REQUESTS" validate:"min=0,max=10" envDefault:"5"`
	Window      time.Duration `env:"WINDOW"       validate:"min=0,max=2m" envDefault:"1m"`
}

// SMTP contains credentials and settings for outgoing email.
type SMTP struct {
	Host     string `env:"HOST"     validate:"required,hostname|ip"`
	Port     uint64 `env:"PORT"     validate:"required,port"`
	Username string `env:"USERNAME" validate:"required"`
	Password string `env:"PASSWORD" validate:"required"`
	From     string `env:"FROM"     validate:"required,email"`
}

// CORS defines cross-origin access rules for the API.
type CORS struct {
	Origins []string `env:"ORIGINS" envSeparator:"," validate:"dive"`
	Methods []string `env:"METHODS" envSeparator:"," validate:"dive,oneof=GET POST PUT DELETE OPTIONS"`
}

// dbConnection stores connection parameters for the database.
type dbConnection struct {
	Driver   string `env:"DRIVER"   validate:"required,oneof=mongo mysql"`
	Host     string `env:"HOST"     validate:"required,hostname|ip"`
	Port     uint64 `env:"PORT"     validate:"required,port"`
	Name     string `env:"NAME"     validate:"required"`
	User     string `env:"USER"     validate:"required"`
	Password string `env:"PASSWORD" validate:"required"`
}

// dbPool stores pool parameters for the database.
type dbPool struct {
	MaxOpen     uint64        `env:"MAX_OPEN"      validate:"min=1,max=200,gtfield=MinOpen" envDefault:"100"`
	MinOpen     uint64        `env:"MIN_OPEN"      validate:"min=0,max=50"                  envDefault:"20"`
	MaxIdleTime time.Duration `env:"MAX_IDLE_TIME" validate:"min=0,max=30m"                 envDefault:"30m"`
}

// dbResilience stores resilience parameters for the database.
type dbResilience struct {
	MaxRetries          int           `env:"MAX_RETRIES"           validate:"min=0,max=10"   envDefault:"3"`
	RetryBackoff        time.Duration `env:"RETRY_BACKOFF"         validate:"min=0,max=30s"  envDefault:"100ms"`
	HealthCheckInterval time.Duration `env:"HEALTH_CHECK_INTERVAL" validate:"min=1m,max=1h"  envDefault:"10m"`
	HealthCheckTimeout  time.Duration `env:"HEALTH_CHECK_TIMEOUT"  validate:"min=1s,max=30s" envDefault:"10s"`
}

// dbTimeout stores timeout parameters for the database.
type dbTimeout struct {
	Connect time.Duration `env:"CONNECT" validate:"min=1s,max=30s" envDefault:"10s"`
	Ping    time.Duration `env:"PING"    validate:"min=1s,max=30s" envDefault:"5s"`
	Close   time.Duration `env:"CLOSE"   validate:"min=1s,max=60s" envDefault:"10s"`
}
