package domain

import (
	"time"

	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeVerifyEmail   TokenType = "verify_email"
	TokenTypePasswordReset TokenType = "password_reset"
)

func (t TokenType) String() string {
	return string(t)
}

func (t TokenType) IsValid() bool {
	return t == TokenTypeVerifyEmail || t == TokenTypePasswordReset
}

type Token struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      TokenType
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func NewToken(userID uuid.UUID, tokenType TokenType, tokenHash string, expiresAt time.Time) (*Token, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}

	if !tokenType.IsValid() {
		return nil, ErrTokenTypeInvalid
	}

	if tokenHash == "" {
		return nil, ErrTokenRequired
	}

	now := time.Now().UTC()
	if !expiresAt.After(now) {
		return nil, ErrTokenExpired
	}

	return &Token{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      tokenType,
		Token:     tokenHash,
		ExpiresAt: expiresAt,
		UsedAt:    nil,
		CreatedAt: now,
	}, nil
}

func (t *Token) IsExpired() bool {
	return !t.ExpiresAt.After(time.Now().UTC())
}

func (t *Token) IsUsed() bool {
	return t.UsedAt != nil
}

func (t *Token) Use() error {
	if t.IsUsed() {
		return ErrTokenUsed
	}

	if t.IsExpired() {
		return ErrTokenExpired
	}

	now := time.Now().UTC()
	t.UsedAt = &now

	return nil
}
