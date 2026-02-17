package domain

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	TokenType TokenType
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func NewToken(userID uuid.UUID, tokenHash string, tokenType TokenType, expiresAt time.Time) (*Token, error) {
	if userID == uuid.Nil {
		return nil, ErrUserIDRequired
	}

	if tokenHash == "" {
		return nil, ErrTokenRequired
	}

	if !tokenType.IsValid() {
		return nil, ErrTokenTypeInvalid
	}

	if !expiresAt.After(time.Now().UTC()) {
		return nil, ErrTokenExpired
	}

	return &Token{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     tokenHash,
		TokenType: tokenType,
		ExpiresAt: expiresAt,
		UsedAt:    nil,
		CreatedAt: time.Now().UTC(),
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
