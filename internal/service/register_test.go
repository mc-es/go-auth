package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-auth/internal/apperror"
	"go-auth/internal/service"
)

var validRegisterReq = &service.RegisterRequest{
	Username: "alice", Email: "alice@example.com", Password: "password123",
	FirstName: "Alice", LastName: "Doe",
}

func TestServiceRegister(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		req      *service.RegisterRequest
		userRepo *mockUserRepo
		hasher   *mockPasswordHasher
		wantErr  bool
		wantCode apperror.Code
	}{
		{
			name:     "nil request",
			req:      nil,
			userRepo: &mockUserRepo{},
			hasher:   &mockPasswordHasher{},
			wantErr:  true,
			wantCode: apperror.ErrCodeInvalidParam,
		},
		{
			name: "invalid username",
			req: &service.RegisterRequest{
				Username: "ab", Email: "u@example.com", Password: "password123", FirstName: "A", LastName: "B",
			},
			userRepo: &mockUserRepo{},
			hasher:   &mockPasswordHasher{},
			wantErr:  true,
			wantCode: apperror.ErrCodeInvalidParam,
		},
		{
			name: "invalid email",
			req: &service.RegisterRequest{
				Username: "alice", Email: "not-an-email", Password: "password123", FirstName: "A", LastName: "B",
			},
			userRepo: &mockUserRepo{},
			hasher:   &mockPasswordHasher{},
			wantErr:  true,
			wantCode: apperror.ErrCodeInvalidParam,
		},
		{
			name:     "username exists",
			req:      validRegisterReq,
			userRepo: &mockUserRepo{existsByUsername: true},
			hasher:   &mockPasswordHasher{},
			wantErr:  true,
			wantCode: apperror.ErrCodeUsernameAlreadyUsed,
		},
		{
			name:     "email exists",
			req:      validRegisterReq,
			userRepo: &mockUserRepo{existsByEmail: true},
			hasher:   &mockPasswordHasher{},
			wantErr:  true,
			wantCode: apperror.ErrCodeEmailAlreadyUsed,
		},
		{
			name:     "hasher error",
			req:      validRegisterReq,
			userRepo: &mockUserRepo{},
			hasher:   &mockPasswordHasher{hashErr: errors.New("hash failed")},
			wantErr:  true,
			wantCode: apperror.ErrCodeInternalServer,
		},
		{
			name:     "save error",
			req:      validRegisterReq,
			userRepo: &mockUserRepo{saveErr: errors.New("db error")},
			hasher:   &mockPasswordHasher{},
			wantErr:  true,
			wantCode: apperror.ErrCodeInternalServer,
		},
		{
			name:     "success",
			req:      validRegisterReq,
			userRepo: &mockUserRepo{},
			hasher:   &mockPasswordHasher{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc, err := newTestServiceWith(testDeps{UserRepo: tt.userRepo, Hasher: tt.hasher})
			require.NoError(t, err)

			got, err := svc.Register(ctx, tt.req)
			if tt.wantErr {
				require.Error(t, err)

				if tt.wantCode != "" {
					assertAppErrorCode(t, err, tt.wantCode)
				}

				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.NotEqual(t, uuid.Nil, got.UserID)

			if tt.name == "success" {
				require.NotNil(t, tt.userRepo.savedUser)
				assert.Equal(t, "alice", tt.userRepo.savedUser.Username.String())
			}
		})
	}
}
