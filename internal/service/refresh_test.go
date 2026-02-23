package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-auth/internal/apperror"
	"go-auth/internal/service"
)

var validRefreshReq = &service.RefreshRequest{RefreshToken: "token", UserAgent: "ua", ClientIP: "1.2.3.4"}

func TestServiceRefresh(t *testing.T) {
	ctx := context.Background()
	user := mustVerifiedUser(t, "alice", "alice@example.com", "$hash")
	userID := user.ID

	const sessionActive, sessionExpired = "active", "expired"

	tests := []struct {
		name        string
		req         *service.RefreshRequest
		sessionKind string
		userRepo    *mockUserRepo
		sessionRepo *mockSessionRepo
		opaque      *mockOpaqueTokenManager
		access      *mockAccessTokenManager
		wantErr     bool
		wantCode    apperror.Code
	}{
		{
			name:     "nil request",
			req:      nil,
			wantErr:  true,
			wantCode: apperror.ErrCodeInvalidParam,
		},
		{
			name:     "empty token",
			req:      &service.RefreshRequest{RefreshToken: "", UserAgent: "ua", ClientIP: "1.2.3.4"},
			wantErr:  true,
			wantCode: apperror.ErrCodeTokenRequired,
		},
		{
			name:        "hash error",
			req:         validRefreshReq,
			opaque:      &mockOpaqueTokenManager{hashErr: errors.New("hash err")},
			sessionRepo: &mockSessionRepo{},
			wantErr:     true,
			wantCode:    apperror.ErrCodeInternalServer,
		},
		{
			name:        "session not found",
			req:         validRefreshReq,
			opaque:      &mockOpaqueTokenManager{hashResult: "h"},
			sessionRepo: &mockSessionRepo{},
			wantErr:     true,
			wantCode:    apperror.ErrCodeSessionNotFound,
		},
		{
			name:        "session not active",
			req:         validRefreshReq,
			sessionKind: sessionExpired,
			opaque:      &mockOpaqueTokenManager{hashResult: "h"},
			sessionRepo: &mockSessionRepo{},
			wantErr:     true,
			wantCode:    apperror.ErrCodeInvalidToken,
		},
		{
			name:        "user not found",
			req:         validRefreshReq,
			sessionKind: sessionActive,
			userRepo:    &mockUserRepo{getByIDUser: nil},
			sessionRepo: &mockSessionRepo{},
			opaque:      &mockOpaqueTokenManager{hashResult: "h", generateToken: "new-rt"},
			access:      &mockAccessTokenManager{},
			wantErr:     true,
			wantCode:    apperror.ErrCodeUserNotFound,
		},
		{
			name:        "success",
			req:         validRefreshReq,
			sessionKind: sessionActive,
			userRepo:    &mockUserRepo{getByIDUser: user},
			sessionRepo: &mockSessionRepo{},
			opaque:      &mockOpaqueTokenManager{hashResult: "h", generateToken: "new-rt"},
			access:      &mockAccessTokenManager{generateToken: "new-at"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sessionRepo := refreshSessionRepo(t, tt.sessionRepo, tt.sessionKind, userID)
			svc, err := newTestServiceWith(testDeps{
				UserRepo:    tt.userRepo,
				SessionRepo: sessionRepo,
				Opaque:      tt.opaque,
				Access:      tt.access,
			})
			require.NoError(t, err)

			got, err := svc.Refresh(ctx, tt.req)
			assertRefreshResult(t, tt.wantErr, tt.wantCode, got, err)
		})
	}
}

func refreshSessionRepo(t *testing.T, base *mockSessionRepo, kind string, userID uuid.UUID) *mockSessionRepo {
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
	}

	return repo
}

func assertRefreshResult(t *testing.T, wantErr bool, wantCode apperror.Code, got *service.RefreshResponse, err error) {
	t.Helper()

	if wantErr {
		require.Error(t, err)

		if wantCode != "" {
			assertAppErrorCode(t, err, wantCode)
		}

		return
	}

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.NotEmpty(t, got.AccessToken)
	assert.NotEmpty(t, got.RefreshToken)
	assert.False(t, got.AccessExpiresAt.IsZero())
	assert.False(t, got.RefreshExpiresAt.IsZero())
}
