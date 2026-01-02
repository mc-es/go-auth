// Package config provides configuration for the application.
//
// Usage:
//
//	cfg, err := config.LoadEnv()
//	if err != nil {
//		panic(err)
//	}
package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

// LoadEnv loads the configuration from the environment variables.
func LoadEnv() (*EnvConfig, error) {
	_ = godotenv.Load()

	cfg := &EnvConfig{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	v := validator.New(validator.WithRequiredStructEnabled())
	if err := v.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// EnvConfig represents the application environment configuration.
type EnvConfig struct {
	AppName string `env:"APP_NAME" validate:"required"`
	Host    string `env:"HOST"     validate:"hostname|ip"    envDefault:"0.0.0.0"`
	Port    uint16 `env:"PORT"     validate:"port"           envDefault:"8080"`
	Env     string `env:"ENV"      validate:"oneof=dev prod" envDefault:"dev"`
}
