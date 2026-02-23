package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-auth/internal/apperror"
	"go-auth/internal/domain"
	"go-auth/internal/service"
)

var validLoginReq = &service.LoginRequest{Login: "alice", Password: "pass", UserAgent: "ua", ClientIP: "1.2.3.4"}

func TestServiceLogin(t *testing.T) {
	ctx := context.Background()
	user := mustVerifiedUser(t, "alice", "alice@example.com", "$hash")
	passHash, _ := domain.NewPasswordFromHash("$hash")
	userWithPass := func() *domain.User {
		u := *user
		u.Password = passHash

		return &u
	}

	tests := []struct {
		name        string
		req         *service.LoginRequest
		userRepo    *mockUserRepo
		sessionRepo *mockSessionRepo
		hasher      *mockPasswordHasher
		opaque      *mockOpaqueTokenManager
		access      *mockAccessTokenManager
		wantErr     bool
		wantCode    apperror.Code
	}{
		{
			name:     "nil request",
			req:      nil,
			userRepo: &mockUserRepo{},
			wantErr:  true,
			wantCode: apperror.ErrCodeInvalidParam,
		},
		{
			name:     "user not found",
			req:      validLoginReq,
			userRepo: &mockUserRepo{},
			wantErr:  true,
			wantCode: apperror.ErrCodeInvalidCredentials,
		},
		{
			name:     "wrong password",
			req:      validLoginReq,
			userRepo: &mockUserRepo{getByUsernameUser: user},
			hasher:   &mockPasswordHasher{compareOk: false},
			wantErr:  true,
			wantCode: apperror.ErrCodeInvalidCredentials,
		},
		{
			name: "user banned",
			req:  validLoginReq,
			userRepo: func() *mockUserRepo {
				u := *user
				u.Password = passHash
				require.NoError(t, u.Ban())

				return &mockUserRepo{getByUsernameUser: &u}
			}(),
			hasher:   &mockPasswordHasher{compareOk: true},
			wantErr:  true,
			wantCode: apperror.ErrCodeUserBlocked,
		},
		{
			name:        "session save error",
			req:         validLoginReq,
			userRepo:    &mockUserRepo{getByUsernameUser: userWithPass()},
			sessionRepo: &mockSessionRepo{saveErr: errors.New("db error")},
			hasher:      &mockPasswordHasher{compareOk: true},
			opaque:      &mockOpaqueTokenManager{},
			access:      &mockAccessTokenManager{},
			wantErr:     true,
			wantCode:    apperror.ErrCodeInternalServer,
		},
		{
			name:        "success",
			req:         validLoginReq,
			userRepo:    &mockUserRepo{getByUsernameUser: userWithPass()},
			sessionRepo: &mockSessionRepo{},
			hasher:      &mockPasswordHasher{compareOk: true},
			opaque:      &mockOpaqueTokenManager{generateToken: "rt", hashResult: "rt-hash"},
			access:      &mockAccessTokenManager{generateToken: "at"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc, err := newTestServiceWith(testDeps{
				UserRepo:    tt.userRepo,
				SessionRepo: tt.sessionRepo,
				Hasher:      tt.hasher,
				Opaque:      tt.opaque,
				Access:      tt.access,
			})
			require.NoError(t, err)

			got, err := svc.Login(ctx, tt.req)
			assertLoginResult(t, tt.wantErr, tt.wantCode, user, got, err)
		})
	}
}

func assertLoginResult(
	t *testing.T,
	wantErr bool,
	wantCode apperror.Code,
	user *domain.User,
	got *service.LoginResponse,
	err error,
) {
	t.Helper()

	if wantErr {
		require.Error(t, err)
		assertAppErrorCode(t, err, wantCode)

		return
	}

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, user.ID, got.UserID)
	assert.NotEmpty(t, got.AccessToken)
	assert.NotEmpty(t, got.RefreshToken)
	assert.False(t, got.AccessExpiresAt.IsZero())
	assert.False(t, got.RefreshExpiresAt.IsZero())
}
