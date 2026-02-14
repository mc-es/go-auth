package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go-auth/internal/config"
)

func validMinimalConfig() *config.Config {
	return &config.Config{
		App: config.App{
			Name: "go-auth",
			Env:  "dev",
		},
		Server: config.Server{
			Host:       "0.0.0.0",
			Port:       8080,
			IdleTO:     60 * time.Second,
			ReadTO:     15 * time.Second,
			WriteTO:    15 * time.Second,
			ShutdownTO: 30 * time.Second,
		},
		CORS: config.CORS{
			Origins: []string{"http://localhost:3000"},
			Methods: []string{"GET", "POST"},
			Headers: []string{"Content-Type"},
			MaxAge:  600,
		},
		RateLimit: config.RateLimit{
			Limit:  100,
			Period: time.Minute,
		},
		Database: config.Database{
			Name:            "testdb",
			Host:            "localhost",
			Port:            5432,
			User:            "testuser",
			Password:        "p@ss",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 30 * time.Minute,
		},
		Security: config.Security{
			JWTSecret:  "01234567890123456789012345678901",
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 48 * time.Hour,
			HashCost:   12,
		},
		SMTP: config.SMTP{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "pass",
			From:     "test@example.com",
		},
	}
}

func TestServerAddr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		host   string
		port   uint16
		expect string
	}{
		{"zero_host_port", "0.0.0.0", 8080, "0.0.0.0:8080"},
		{"localhost", "localhost", 3000, "localhost:3000"},
		{"ipv4", "192.168.1.1", 443, "192.168.1.1:443"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := validMinimalConfig()
			cfg.Server.Host = tt.host
			cfg.Server.Port = tt.port
			assert.Equal(t, tt.expect, cfg.ServerAddr())
		})
	}
}

func TestSMTPAddr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		host   string
		port   uint16
		expect string
	}{
		{"smtp_default", "smtp.example.com", 587, "smtp.example.com:587"},
		{"smtp_25", "mail.local", 25, "mail.local:25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := validMinimalConfig()
			cfg.SMTP.Host = tt.host
			cfg.SMTP.Port = tt.port
			assert.Equal(t, tt.expect, cfg.SMTPAddr())
		})
	}
}

func TestDatabaseURL(t *testing.T) {
	t.Parallel()

	cfg := validMinimalConfig()
	got := cfg.DatabaseURL()
	assert.Contains(t, got, "postgres://")
	assert.Contains(t, got, "testuser")
	assert.Contains(t, got, "localhost:5432")
	assert.Contains(t, got, "testdb")
	assert.Contains(t, got, "sslmode=disable")
}

func TestJWTKey(t *testing.T) {
	t.Parallel()

	cfg := validMinimalConfig()
	secret := "01234567890123456789012345678901"
	cfg.Security.JWTSecret = secret
	got := cfg.JWTKey()
	assert.Equal(t, []byte(secret), got)
	assert.Len(t, got, 32)
}

func TestIsProduction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		env    string
		expect bool
	}{
		{"prod", true},
		{"dev", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			t.Parallel()

			cfg := validMinimalConfig()
			cfg.App.Env = tt.env
			assert.Equal(t, tt.expect, cfg.IsProduction())
		})
	}
}

func TestIsDevelopment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		env    string
		expect bool
	}{
		{"dev", true},
		{"prod", false},
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			t.Parallel()

			cfg := validMinimalConfig()
			cfg.App.Env = tt.env
			assert.Equal(t, tt.expect, cfg.IsDevelopment())
		})
	}
}
