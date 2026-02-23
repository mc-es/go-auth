package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"go-auth/internal/domain"
)

const (
	sessionTestToken     = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	sessionTestUserAgent = "Mozilla/5.0"
	sessionTestClientIP  = "192.168.1.1"
)

func mustSession(t *testing.T, opts ...func(*domain.Session)) *domain.Session {
	t.Helper()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	s, err := domain.NewSession(userID, sessionTestToken, sessionTestUserAgent, sessionTestClientIP, expiresAt)
	assert.NoError(t, err)

	for _, fn := range opts {
		fn(s)
	}

	return s
}

func TestNewSession(t *testing.T) {
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	tests := []struct {
		name      string
		userID    uuid.UUID
		token     string
		userAgent string
		clientIP  string
		expiresAt time.Time
		wantErr   error
	}{
		{
			name:      "valid",
			userID:    userID,
			token:     sessionTestToken,
			userAgent: sessionTestUserAgent,
			clientIP:  sessionTestClientIP,
			expiresAt: expiresAt,
			wantErr:   nil,
		},
		{
			name:      "nil user id",
			userID:    uuid.Nil,
			token:     sessionTestToken,
			userAgent: sessionTestUserAgent,
			clientIP:  sessionTestClientIP,
			expiresAt: expiresAt,
			wantErr:   domain.ErrUserIDRequired,
		},
		{
			name:      "empty token",
			userID:    userID,
			token:     "",
			userAgent: sessionTestUserAgent,
			clientIP:  sessionTestClientIP,
			expiresAt: expiresAt,
			wantErr:   domain.ErrTokenRequired,
		},
		{
			name:      "expired at in past",
			userID:    userID,
			token:     sessionTestToken,
			userAgent: sessionTestUserAgent,
			clientIP:  sessionTestClientIP,
			expiresAt: time.Now().UTC().Add(-time.Hour),
			wantErr:   domain.ErrSessionExpired,
		},
		{
			name:      "expires at now",
			userID:    userID,
			token:     sessionTestToken,
			userAgent: sessionTestUserAgent,
			clientIP:  sessionTestClientIP,
			expiresAt: time.Now().UTC(),
			wantErr:   domain.ErrSessionExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := domain.NewSession(tt.userID, tt.token, tt.userAgent, tt.clientIP, tt.expiresAt)
			assert.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr != nil {
				assert.Nil(t, got)

				return
			}

			assert.NotEqual(t, uuid.Nil, got.ID)
			assert.Equal(t, tt.userID, got.UserID)
			assert.Equal(t, tt.token, got.Token)
			assert.Equal(t, tt.userAgent, got.UserAgent)
			assert.Equal(t, tt.clientIP, got.ClientIP)
			assert.WithinDuration(t, expiresAt, got.ExpiresAt, time.Second)
			assert.Nil(t, got.RevokedAt)
			assert.False(t, got.CreatedAt.IsZero())
			assert.False(t, got.UpdatedAt.IsZero())
		})
	}
}

func TestSessionIsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		s := mustSession(t)
		assert.False(t, s.IsExpired())
	})

	t.Run("expired", func(t *testing.T) {
		s := mustSession(t)
		s.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		assert.True(t, s.IsExpired())
	})

	t.Run("zero time", func(t *testing.T) {
		var s domain.Session
		assert.True(t, s.IsExpired())
	})
}

func TestSessionIsRevoked(t *testing.T) {
	t.Run("not revoked", func(t *testing.T) {
		s := mustSession(t)
		assert.False(t, s.IsRevoked())
	})

	t.Run("revoked", func(t *testing.T) {
		s := mustSession(t)
		now := time.Now().UTC()
		s.RevokedAt = &now
		assert.True(t, s.IsRevoked())
	})
}

func TestSessionIsActive(t *testing.T) {
	t.Run("active", func(t *testing.T) {
		s := mustSession(t)
		assert.True(t, s.IsActive())
	})

	t.Run("expired", func(t *testing.T) {
		s := mustSession(t)
		s.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		assert.False(t, s.IsActive())
	})

	t.Run("revoked", func(t *testing.T) {
		s := mustSession(t)
		now := time.Now().UTC()
		s.RevokedAt = &now
		assert.False(t, s.IsActive())
	})

	t.Run("expired and revoked", func(t *testing.T) {
		s := mustSession(t)
		s.ExpiresAt = time.Now().UTC().Add(-time.Minute)
		now := time.Now().UTC()
		s.RevokedAt = &now
		assert.False(t, s.IsActive())
	})
}

func TestSessionRevoke(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		s := mustSession(t)
		assert.NoError(t, s.Revoke())
		assert.True(t, s.IsRevoked())
		assert.NotNil(t, s.RevokedAt)
	})

	t.Run("already revoked", func(t *testing.T) {
		s := mustSession(t)
		now := time.Now().UTC()
		s.RevokedAt = &now
		assert.ErrorIs(t, s.Revoke(), domain.ErrSessionRevoked)
	})
}

func TestSessionRotate(t *testing.T) {
	newToken := "new-jwt-token"
	newExpiresAt := time.Now().UTC().Add(48 * time.Hour)

	t.Run("ok", func(t *testing.T) {
		s := mustSession(t)
		oldID := s.ID

		newS, err := s.Rotate(newToken, newExpiresAt, "", "")
		assert.NoError(t, err)
		assert.NotNil(t, newS)

		assert.True(t, s.IsRevoked())
		assert.NotEqual(t, oldID, newS.ID)
		assert.Equal(t, s.UserID, newS.UserID)
		assert.Equal(t, newToken, newS.Token)
		assert.Equal(t, s.UserAgent, newS.UserAgent)
		assert.Equal(t, s.ClientIP, newS.ClientIP)
		assert.True(t, newS.ExpiresAt.After(time.Now().UTC()))
		assert.Nil(t, newS.RevokedAt)
	})

	t.Run("revoked session", func(t *testing.T) {
		s := mustSession(t)
		now := time.Now().UTC()
		s.RevokedAt = &now

		newS, err := s.Rotate(newToken, newExpiresAt, "", "")
		assert.ErrorIs(t, err, domain.ErrSessionRevoked)
		assert.Nil(t, newS)
	})

	t.Run("expired session", func(t *testing.T) {
		s := mustSession(t)
		s.ExpiresAt = time.Now().UTC().Add(-time.Minute)

		newS, err := s.Rotate(newToken, newExpiresAt, "", "")
		assert.ErrorIs(t, err, domain.ErrSessionExpired)
		assert.Nil(t, newS)
	})

	t.Run("empty new token", func(t *testing.T) {
		s := mustSession(t)

		newS, err := s.Rotate("", newExpiresAt, "", "")
		assert.ErrorIs(t, err, domain.ErrTokenRequired)
		assert.Nil(t, newS)
	})

	t.Run("new expires at in past", func(t *testing.T) {
		s := mustSession(t)

		newS, err := s.Rotate(newToken, time.Now().UTC().Add(-time.Hour), "", "")
		assert.ErrorIs(t, err, domain.ErrSessionExpired)
		assert.Nil(t, newS)
	})

	t.Run("new session uses provided userAgent and clientIP", func(t *testing.T) {
		s := mustSession(t)
		newUA := "Mozilla/5.0 (refresh)"
		newIP := "192.168.1.2"

		newS, err := s.Rotate(newToken, newExpiresAt, newUA, newIP)
		assert.NoError(t, err)
		assert.NotNil(t, newS)
		assert.Equal(t, newUA, newS.UserAgent)
		assert.Equal(t, newIP, newS.ClientIP)
	})
}
