package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	UserAgent string
	ClientIP  string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSession(userID uuid.UUID, tokenHash, userAgent, clientIP string, expiresAt time.Time) (*Session, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}

	if tokenHash == "" {
		return nil, ErrTokenRequired
	}

	now := time.Now().UTC()
	if !expiresAt.After(now) {
		return nil, ErrSessionExpired
	}

	return &Session{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     tokenHash,
		UserAgent: userAgent,
		ClientIP:  clientIP,
		ExpiresAt: expiresAt,
		RevokedAt: nil,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Session) IsExpired() bool {
	return !s.ExpiresAt.After(time.Now().UTC())
}

func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *Session) IsActive() bool {
	return !s.IsExpired() && !s.IsRevoked()
}

func (s *Session) Revoke() error {
	if s.IsRevoked() {
		return ErrSessionRevoked
	}

	now := time.Now().UTC()
	s.RevokedAt = &now
	s.touch()

	return nil
}

func (s *Session) Rotate(newToken string, newExpiresAt time.Time, newUserAgent, newClientIP string) (*Session, error) {
	if s.IsRevoked() {
		return nil, ErrSessionRevoked
	}

	if s.IsExpired() {
		return nil, ErrSessionExpired
	}

	if newToken == "" {
		return nil, ErrTokenRequired
	}

	now := time.Now().UTC()
	if !newExpiresAt.After(now) {
		return nil, ErrSessionExpired
	}

	s.RevokedAt = &now
	s.touch()

	userAgent := newUserAgent
	if userAgent == "" {
		userAgent = s.UserAgent
	}

	clientIP := newClientIP
	if clientIP == "" {
		clientIP = s.ClientIP
	}

	newSession := &Session{
		ID:        uuid.New(),
		UserID:    s.UserID,
		Token:     newToken,
		UserAgent: userAgent,
		ClientIP:  clientIP,
		ExpiresAt: newExpiresAt,
		RevokedAt: nil,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return newSession, nil
}

func (s *Session) touch() {
	s.UpdatedAt = time.Now().UTC()
}
