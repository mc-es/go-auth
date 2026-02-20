package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go-auth/internal/domain"
)

const (
	usernameAlice = "alice"
	usernameBob   = "bob42"
	usernameUser  = "valid_user"
)

func TestUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "Valid simple username",
			input: "alice",
			want:  usernameAlice,
		},
		{
			name:  "Valid username with uppercase and spaces",
			input: "  Alice  ",
			want:  usernameAlice,
		},
		{
			name:  "Valid username with digits",
			input: "bob42",
			want:  usernameBob,
		},
		{
			name:  "Valid username with underscore",
			input: "valid_user",
			want:  usernameUser,
		},
		{
			name:  "Valid min length (3 chars)",
			input: "abc",
			want:  "abc",
		},
		{
			name:  "Valid max length (20 chars)",
			input: "a1234567890123456789",
			want:  "a1234567890123456789",
		},
		{
			name:    "Invalid empty",
			input:   "",
			wantErr: domain.ErrUsernameRequired,
		},
		{
			name:    "Invalid whitespace only",
			input:   "   ",
			wantErr: domain.ErrUsernameRequired,
		},
		{
			name:    "Invalid too short (1 char)",
			input:   "a",
			wantErr: domain.ErrUsernameTooShort,
		},
		{
			name:    "Invalid too long (21 chars)",
			input:   "a12345678901234567890",
			wantErr: domain.ErrUsernameTooLong,
		},
		{
			name:    "Invalid starts with digit",
			input:   "1alice",
			wantErr: domain.ErrUsernameInvalid,
		},
		{
			name:    "Invalid character hyphen",
			input:   "alice-bob",
			wantErr: domain.ErrUsernameInvalid,
		},
		{
			name:    "Invalid character dot",
			input:   "alice.bob",
			wantErr: domain.ErrUsernameInvalid,
		},
		{
			name:    "Reserved username admin",
			input:   "admin",
			wantErr: domain.ErrUsernameInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := domain.NewUsername(tt.input)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestUsernameIsZero(t *testing.T) {
	var zero domain.Username
	assert.True(t, zero.IsZero())

	username, _ := domain.NewUsername(usernameUser)
	assert.False(t, username.IsZero())
}

func TestUsernameValue(t *testing.T) {
	var zero domain.Username

	val, err := zero.Value()
	assert.Empty(t, val)
	assert.ErrorIs(t, err, domain.ErrUsernameRequired)

	username, _ := domain.NewUsername(usernameUser)
	val, err = username.Value()
	assert.NoError(t, err)
	assert.NotEmpty(t, val)
	assert.Equal(t, usernameUser, val)
}

func TestUsernameScan(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr error
	}{
		{"nil", nil, "", nil},
		{"string", usernameAlice, usernameAlice, nil},
		{"bytes", []byte(usernameBob), usernameBob, nil},
		{"invalid type", 123, "", domain.ErrUsernameScan},
		{"invalid username string", "ab", "", domain.ErrUsernameTooShort},
		{"empty string", "", "", domain.ErrUsernameRequired},
		{"reserved username", "admin", "", domain.ErrUsernameInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var u domain.Username

			err := u.Scan(tt.input)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, u.String())
		})
	}
}

func TestUsernameMarshalText(t *testing.T) {
	var zero domain.Username

	data, err := zero.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), data)

	username, _ := domain.NewUsername(usernameUser)
	data, err = username.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte(usernameUser), data)
}

func TestUsernameUnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    string
		wantErr error
	}{
		{"valid", []byte(usernameAlice), usernameAlice, nil},
		{"empty", []byte(""), "", nil},
		{"invalid too short", []byte("ab"), "", domain.ErrUsernameTooShort},
		{"invalid reserved", []byte("root"), "", domain.ErrUsernameInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var u domain.Username

			err := u.UnmarshalText(tt.input)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, u.String())
		})
	}
}
