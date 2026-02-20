package security_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-auth/internal/domain"
	"go-auth/internal/security"
)

const (
	userID        = "550e8400-e29b-41d4-a716-446655440000"
	jwtTestSecret = "01234567890123456789012345678901"
	jwtTestIssuer = "go-auth-test"
)

func TestNewJWT(t *testing.T) {
	tests := []struct {
		name       string
		secret     string
		issuer     string
		accessTTL  time.Duration
		refreshTTL time.Duration
		wantErr    error
	}{
		{
			name:       "ok",
			secret:     jwtTestSecret,
			issuer:     jwtTestIssuer,
			accessTTL:  time.Hour,
			refreshTTL: 24 * time.Hour,
			wantErr:    nil,
		},
		{
			name:       "empty secret",
			secret:     "",
			issuer:     jwtTestIssuer,
			accessTTL:  time.Hour,
			refreshTTL: 24 * time.Hour,
			wantErr:    domain.ErrTokenSecretRequired,
		},
		{
			name:       "zero access TTL",
			secret:     jwtTestSecret,
			issuer:     jwtTestIssuer,
			accessTTL:  0,
			refreshTTL: 24 * time.Hour,
			wantErr:    domain.ErrTokenAccessTTLRequired,
		},
		{
			name:       "negative access TTL",
			secret:     jwtTestSecret,
			issuer:     jwtTestIssuer,
			accessTTL:  -time.Minute,
			refreshTTL: 24 * time.Hour,
			wantErr:    domain.ErrTokenAccessTTLRequired,
		},
		{
			name:       "zero refresh TTL",
			secret:     jwtTestSecret,
			issuer:     jwtTestIssuer,
			accessTTL:  time.Hour,
			refreshTTL: 0,
			wantErr:    domain.ErrTokenRefreshTTLRequired,
		},
		{
			name:       "negative refresh TTL",
			secret:     jwtTestSecret,
			issuer:     jwtTestIssuer,
			accessTTL:  time.Hour,
			refreshTTL: -time.Hour,
			wantErr:    domain.ErrTokenRefreshTTLRequired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := security.NewJWT(tt.secret, tt.issuer, tt.accessTTL, tt.refreshTTL)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)

				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}

func TestJWTRoundTrip(t *testing.T) {
	userID := uuid.MustParse(userID)
	role, err := domain.NewRole(domain.RoleUser)
	require.NoError(t, err)

	claims := domain.Claims{
		UserID: userID,
		Role:   role,
		Type:   domain.TokenTypeAccess,
	}

	m, err := security.NewJWT(jwtTestSecret, jwtTestIssuer, time.Hour, 24*time.Hour)
	require.NoError(t, err)
	token, err := m.Generate(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	got, err := m.Validate(token)
	require.NoError(t, err)
	assert.Equal(t, claims.UserID, got.UserID)
	assert.Equal(t, claims.Role.String(), got.Role.String())
	assert.Equal(t, claims.Type, got.Type)
}

func TestJWTInvalidToken(t *testing.T) {
	m, err := security.NewJWT(jwtTestSecret, jwtTestIssuer, time.Hour, 24*time.Hour)
	require.NoError(t, err)

	_, err = m.Validate("invalid.jwt.token")
	assert.Error(t, err)

	_, err = m.Validate("")
	assert.Error(t, err)
}

func TestJWTWrongSecret(t *testing.T) {
	userID := uuid.MustParse(userID)
	role, _ := domain.NewRole(domain.RoleAdmin)
	claims := domain.Claims{
		UserID: userID,
		Role:   role,
		Type:   domain.TokenTypeAccess,
	}

	gen, err := security.NewJWT(jwtTestSecret, jwtTestIssuer, time.Hour, 24*time.Hour)
	require.NoError(t, err)
	token, err := gen.Generate(claims)
	require.NoError(t, err)

	validator, err := security.NewJWT("different-secret-32-bytes-long!!!!!", jwtTestIssuer, time.Hour, 24*time.Hour)
	require.NoError(t, err)
	_, err = validator.Validate(token)
	assert.Error(t, err)
}

func TestJWTExpiredToken(t *testing.T) {
	userID := uuid.MustParse(userID)
	role, _ := domain.NewRole(domain.RoleUser)
	claims := domain.Claims{
		UserID: userID,
		Role:   role,
		Type:   domain.TokenTypeAccess,
	}

	m, err := security.NewJWT(jwtTestSecret, jwtTestIssuer, time.Millisecond, 24*time.Hour)
	require.NoError(t, err)
	token, err := m.Generate(claims)
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)

	_, err = m.Validate(token)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTokenExpired)
}
