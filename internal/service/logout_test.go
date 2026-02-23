package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"go-auth/internal/apperror"
)

func TestServiceLogout(t *testing.T) {
	ctx := context.Background()
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	const sessionNone, sessionActive, sessionExpired, sessionRevoked = "", "active", "expired", "revoked"

	tests := []struct {
		name         string
		refreshToken string
		sessionKind  string
		opaque       *mockOpaqueTokenManager
		sessionRepo  *mockSessionRepo
		wantErr      bool
		wantCode     apperror.Code
	}{
		{
			name:         "empty token",
			refreshToken: "",
			sessionRepo:  &mockSessionRepo{},
			wantErr:      true,
			wantCode:     apperror.ErrCodeInvalidParam,
		},
		{
			name:         "hash error",
			refreshToken: "token",
			opaque:       &mockOpaqueTokenManager{hashErr: errors.New("hash err")},
			sessionRepo:  &mockSessionRepo{},
			wantErr:      true,
			wantCode:     apperror.ErrCodeInternalServer,
		},
		{
			name:         "session not found",
			refreshToken: "token",
			sessionKind:  sessionNone,
			opaque:       &mockOpaqueTokenManager{hashResult: "hash"},
			sessionRepo:  &mockSessionRepo{},
			wantErr:      true,
			wantCode:     apperror.ErrCodeSessionNotFound,
		},
		{
			name:         "session expired",
			refreshToken: "token",
			sessionKind:  sessionExpired,
			opaque:       &mockOpaqueTokenManager{hashResult: "hash"},
			sessionRepo:  &mockSessionRepo{},
			wantErr:      true,
			wantCode:     apperror.ErrCodeInvalidToken,
		},
		{
			name:         "session revoked",
			refreshToken: "token",
			sessionKind:  sessionRevoked,
			opaque:       &mockOpaqueTokenManager{hashResult: "hash"},
			sessionRepo:  &mockSessionRepo{},
			wantErr:      true,
			wantCode:     apperror.ErrCodeInvalidToken,
		},
		{
			name:         "update error",
			refreshToken: "token",
			sessionKind:  sessionActive,
			opaque:       &mockOpaqueTokenManager{hashResult: "hash"},
			sessionRepo:  &mockSessionRepo{updateErr: errors.New("db error")},
			wantErr:      true,
			wantCode:     apperror.ErrCodeInternalServer,
		},
		{
			name:         "success",
			refreshToken: "token",
			sessionKind:  sessionActive,
			opaque:       &mockOpaqueTokenManager{hashResult: "hash"},
			sessionRepo:  &mockSessionRepo{},
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sessionRepo := logoutSessionRepo(t, tt.sessionRepo, tt.sessionKind, userID)
			svc, err := newTestServiceWith(testDeps{
				SessionRepo: sessionRepo,
				Opaque:      tt.opaque,
			})
			require.NoError(t, err)

			err = svc.Logout(ctx, tt.refreshToken)
			assertLogoutResult(t, tt.wantErr, tt.wantCode, err)
		})
	}
}

func logoutSessionRepo(t *testing.T, base *mockSessionRepo, kind string, userID uuid.UUID) *mockSessionRepo {
	t.Helper()

	repo := &mockSessionRepo{}
	if base != nil {
		*repo = *base
	}

	switch kind {
	case "active":
		repo.getByToken = mustSession(t, userID, 24*time.Hour, false)
	case "expired":
		repo.getByToken = mustSession(t, userID, -time.Hour, false)
	case "revoked":
		repo.getByToken = mustSession(t, userID, 24*time.Hour, true)
	}

	return repo
}

func assertLogoutResult(t *testing.T, wantErr bool, wantCode apperror.Code, err error) {
	t.Helper()

	if wantErr {
		require.Error(t, err)

		if wantCode != "" {
			assertAppErrorCode(t, err, wantCode)
		}

		return
	}

	require.NoError(t, err)
}
