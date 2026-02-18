package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"go-auth/internal/domain"
)

const (
	tokenTestHash = "argon2id$v=19$m=65536,t=3,p=2$salt$hash"
)

func mustToken(t *testing.T, opts ...func(*domain.Token)) *domain.Token {
	t.Helper()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	tok, err := domain.NewToken(userID, tokenTestHash, domain.TokenTypeVerifyEmail, expiresAt)
	assert.NoError(t, err)

	for _, fn := range opts {
		fn(tok)
	}

	return tok
}

func TestNewToken(t *testing.T) {
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	tests := []struct {
		name      string
		userID    uuid.UUID
		tokenHash string
		tokenType domain.TokenType
		expiresAt time.Time
		wantErr   error
	}{
		{
			name:      "valid verify_email",
			userID:    userID,
			tokenHash: tokenTestHash,
			tokenType: domain.TokenTypeVerifyEmail,
			expiresAt: expiresAt,
			wantErr:   nil,
		},
		{
			name:      "valid password_reset",
			userID:    userID,
			tokenHash: tokenTestHash,
			tokenType: domain.TokenTypePasswordReset,
			expiresAt: expiresAt,
			wantErr:   nil,
		},
		{
			name:      "valid magic_link",
			userID:    userID,
			tokenHash: tokenTestHash,
			tokenType: domain.TokenTypeMagicLink,
			expiresAt: expiresAt,
			wantErr:   nil,
		},
		{
			name:      "nil user id",
			userID:    uuid.Nil,
			tokenHash: tokenTestHash,
			tokenType: domain.TokenTypeVerifyEmail,
			expiresAt: expiresAt,
			wantErr:   domain.ErrUserIDRequired,
		},
		{
			name:      "empty token hash",
			userID:    userID,
			tokenHash: "",
			tokenType: domain.TokenTypeVerifyEmail,
			expiresAt: expiresAt,
			wantErr:   domain.ErrTokenRequired,
		},
		{
			name:      "invalid token type",
			userID:    userID,
			tokenHash: tokenTestHash,
			tokenType: domain.TokenType("invalid"),
			expiresAt: expiresAt,
			wantErr:   domain.ErrTokenTypeInvalid,
		},
		{
			name:      "expires at in past",
			userID:    userID,
			tokenHash: tokenTestHash,
			tokenType: domain.TokenTypeVerifyEmail,
			expiresAt: time.Now().UTC().Add(-time.Hour),
			wantErr:   domain.ErrTokenExpired,
		},
		{
			name:      "expires at now",
			userID:    userID,
			tokenHash: tokenTestHash,
			tokenType: domain.TokenTypeVerifyEmail,
			expiresAt: time.Now().UTC(),
			wantErr:   domain.ErrTokenExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := domain.NewToken(tt.userID, tt.tokenHash, tt.tokenType, tt.expiresAt)
			assert.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr != nil {
				assert.Nil(t, got)

				return
			}

			assert.NotEqual(t, uuid.Nil, got.ID)
			assert.Equal(t, tt.userID, got.UserID)
			assert.Equal(t, tt.tokenHash, got.Token)
			assert.Equal(t, tt.tokenType, got.TokenType)
			assert.True(t, got.ExpiresAt.After(time.Now().UTC()))
			assert.Nil(t, got.UsedAt)
			assert.False(t, got.CreatedAt.IsZero())
		})
	}
}

func TestTokenIsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		tok := mustToken(t)
		assert.False(t, tok.IsExpired())
	})

	t.Run("expired", func(t *testing.T) {
		tok := mustToken(t)
		tok.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		assert.True(t, tok.IsExpired())
	})

	t.Run("zero time", func(t *testing.T) {
		var tok domain.Token
		assert.True(t, tok.IsExpired())
	})
}

func TestTokenIsUsed(t *testing.T) {
	t.Run("not used", func(t *testing.T) {
		tok := mustToken(t)
		assert.False(t, tok.IsUsed())
	})

	t.Run("used", func(t *testing.T) {
		tok := mustToken(t)
		now := time.Now().UTC()
		tok.UsedAt = &now
		assert.True(t, tok.IsUsed())
	})
}

func TestTokenUse(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		tok := mustToken(t)
		assert.NoError(t, tok.Use())
		assert.True(t, tok.IsUsed())
		assert.NotNil(t, tok.UsedAt)
	})

	t.Run("already used", func(t *testing.T) {
		tok := mustToken(t)
		now := time.Now().UTC()
		tok.UsedAt = &now
		assert.ErrorIs(t, tok.Use(), domain.ErrTokenUsed)
	})

	t.Run("expired", func(t *testing.T) {
		tok := mustToken(t)
		tok.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		assert.ErrorIs(t, tok.Use(), domain.ErrTokenExpired)
	})
}
