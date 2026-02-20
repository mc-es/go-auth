package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go-auth/internal/domain"
)

func TestRole(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "Valid user",
			input: "user",
			want:  "user",
		},
		{
			name:  "Valid admin",
			input: "admin",
			want:  "admin",
		},
		{
			name:  "Valid with uppercase and spaces",
			input: "  Admin  ",
			want:  "admin",
		},
		{
			name:    "Invalid empty",
			input:   "",
			wantErr: domain.ErrRoleRequired,
		},
		{
			name:    "Invalid whitespace only",
			input:   "   ",
			wantErr: domain.ErrRoleRequired,
		},
		{
			name:    "Invalid unknown role",
			input:   "moderator",
			wantErr: domain.ErrRoleInvalid,
		},
		{
			name:    "Invalid partial match",
			input:   "super",
			wantErr: domain.ErrRoleInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := domain.NewRole(tt.input)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestRoleIsZero(t *testing.T) {
	var zero domain.Role
	assert.True(t, zero.IsZero())

	r, _ := domain.NewRole(domain.RoleUser)
	assert.False(t, r.IsZero())
}

func TestRoleHasPermission(t *testing.T) {
	user, _ := domain.NewRole(domain.RoleUser)
	admin, _ := domain.NewRole(domain.RoleAdmin)
	superadmin, _ := domain.NewRole(domain.RoleSuperAdmin)

	tests := []struct {
		name    string
		role    domain.Role
		perm    domain.Permission
		wantHas bool
	}{
		{"user has user:read", user, domain.PermUserRead, true},
		{"user lacks user:write", user, domain.PermUserWrite, false},
		{"user lacks user:ban", user, domain.PermUserBan, false},
		{"user lacks user:delete", user, domain.PermUserDelete, false},

		{"admin has user:read", admin, domain.PermUserRead, true},
		{"admin has user:write", admin, domain.PermUserWrite, true},
		{"admin has user:ban", admin, domain.PermUserBan, true},
		{"admin lacks user:delete", admin, domain.PermUserDelete, false},

		{"superadmin has user:read", superadmin, domain.PermUserRead, true},
		{"superadmin has user:write", superadmin, domain.PermUserWrite, true},
		{"superadmin has user:ban", superadmin, domain.PermUserBan, true},
		{"superadmin has user:delete", superadmin, domain.PermUserDelete, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.role.HasPermission(tt.perm)
			assert.Equal(t, tt.wantHas, got)
		})
	}
}

func TestRoleValue(t *testing.T) {
	var zero domain.Role

	val, err := zero.Value()
	assert.Empty(t, val)
	assert.ErrorIs(t, err, domain.ErrRoleRequired)

	user, _ := domain.NewRole(domain.RoleUser)

	val, err = user.Value()
	assert.NoError(t, err)
	assert.Equal(t, domain.RoleUser, val)
}

func TestRoleScan(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr error
	}{
		{"nil", nil, "", nil},
		{"string user", domain.RoleUser, domain.RoleUser, nil},
		{"string admin", domain.RoleAdmin, domain.RoleAdmin, nil},
		{"string superadmin", domain.RoleSuperAdmin, domain.RoleSuperAdmin, nil},
		{"bytes", []byte(domain.RoleSuperAdmin), domain.RoleSuperAdmin, nil},
		{"invalid type", 123, "", domain.ErrRoleScan},
		{"empty string", "", "", domain.ErrRoleRequired},
		{"invalid role string", "invalid", "", domain.ErrRoleInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var r domain.Role

			err := r.Scan(tt.input)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, r.String())
		})
	}
}

func TestRoleMarshalText(t *testing.T) {
	var zero domain.Role

	data, err := zero.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), data)

	admin, _ := domain.NewRole(domain.RoleAdmin)
	data, err = admin.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte(domain.RoleAdmin), data)
}

func TestRoleUnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    string
		wantErr error
	}{
		{"valid user", []byte(domain.RoleUser), domain.RoleUser, nil},
		{"valid admin", []byte(domain.RoleAdmin), domain.RoleAdmin, nil},
		{"valid superadmin", []byte(domain.RoleSuperAdmin), domain.RoleSuperAdmin, nil},
		{"empty", []byte(""), "", nil},
		{"invalid role", []byte("invalid"), "", domain.ErrRoleInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var r domain.Role

			err := r.UnmarshalText(tt.input)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, r.String())
		})
	}
}
