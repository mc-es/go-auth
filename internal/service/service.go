package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"go-auth/internal/domain"
)

type Service interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, req *RefreshRequest) (*RefreshResponse, error)
}

type RegisterRequest struct {
	Username  string
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type RegisterResponse struct {
	UserID uuid.UUID
}

type LoginRequest struct {
	Login     string
	Password  string
	UserAgent string
	ClientIP  string
}

type LoginResponse struct {
	UserID           uuid.UUID
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type RefreshRequest struct {
	RefreshToken string
	UserAgent    string
	ClientIP     string
}

type RefreshResponse struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type Config struct {
	UserRepo           domain.UserRepository
	SessionRepo        domain.SessionRepository
	PasswordHasher     domain.PasswordHasher
	OpaqueTokenManager domain.OpaqueTokenManager
	AccessTokenManager domain.AccessTokenManager
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
}

type service struct {
	userRepo           domain.UserRepository
	sessionRepo        domain.SessionRepository
	passwordHasher     domain.PasswordHasher
	opaqueTokenManager domain.OpaqueTokenManager
	accessTokenManager domain.AccessTokenManager
	accessTokenTTL     time.Duration
	refreshTokenTTL    time.Duration
}

func NewService(cfg *Config) (Service, error) {
	if cfg == nil {
		return nil, errors.New("service config is required")
	}

	if cfg.AccessTokenTTL <= 0 {
		return nil, errors.New("access token TTL must be positive")
	}

	if cfg.RefreshTokenTTL <= 0 {
		return nil, errors.New("refresh token TTL must be positive")
	}

	if cfg.UserRepo == nil {
		return nil, errors.New("user repository is required")
	}

	if cfg.SessionRepo == nil {
		return nil, errors.New("session repository is required")
	}

	if cfg.PasswordHasher == nil {
		return nil, errors.New("password hasher is required")
	}

	if cfg.OpaqueTokenManager == nil {
		return nil, errors.New("opaque token manager is required")
	}

	if cfg.AccessTokenManager == nil {
		return nil, errors.New("access token manager is required")
	}

	return &service{
		userRepo:           cfg.UserRepo,
		sessionRepo:        cfg.SessionRepo,
		passwordHasher:     cfg.PasswordHasher,
		opaqueTokenManager: cfg.OpaqueTokenManager,
		accessTokenManager: cfg.AccessTokenManager,
		accessTokenTTL:     cfg.AccessTokenTTL,
		refreshTokenTTL:    cfg.RefreshTokenTTL,
	}, nil
}
