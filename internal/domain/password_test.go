package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go-auth/internal/domain"
)

const hashArgon2 = "$argon2id$v=19$m=65536,t=3,p=2$salt$hash"

func TestPassword(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr error
	}{
		{
			name:    "Empty hash",
			hash:    "",
			wantErr: domain.ErrPasswordRequired,
		},
		{
			name:    "Valid hash",
			hash:    hashArgon2,
			wantErr: nil,
		},
		{
			name:    "Single character hash",
			hash:    "x",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := domain.NewPasswordFromHash(tt.hash)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestPasswordIsZero(t *testing.T) {
	var zero domain.Password
	assert.True(t, zero.IsZero())

	password, _ := domain.NewPasswordFromHash(hashArgon2)
	assert.False(t, password.IsZero())
}

func TestPasswordString(t *testing.T) {
	var zero domain.Password
	assert.Equal(t, "*****", zero.String())

	password, _ := domain.NewPasswordFromHash(hashArgon2)
	assert.Equal(t, "*****", password.String())
}

func TestPasswordValue(t *testing.T) {
	var zero domain.Password

	val, err := zero.Value()
	assert.Empty(t, val)
	assert.ErrorIs(t, err, domain.ErrPasswordRequired)

	password, _ := domain.NewPasswordFromHash(hashArgon2)
	val, err = password.Value()
	assert.NoError(t, err)
	assert.NotEmpty(t, val)
	assert.Equal(t, hashArgon2, val)
}

func TestPasswordScan(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr error
	}{
		{"nil", nil, nil},
		{"string", hashArgon2, nil},
		{"bytes", []byte(hashArgon2), nil},
		{"invalid type", 123, domain.ErrPasswordScan},
		{"empty string", "", domain.ErrPasswordRequired},
		{"empty bytes", []byte(""), domain.ErrPasswordRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var p domain.Password

			err := p.Scan(tt.input)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestPasswordMarshalText(t *testing.T) {
	var zero domain.Password

	data, err := zero.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte("*****"), data)

	password, _ := domain.NewPasswordFromHash(hashArgon2)
	data, err = password.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte("*****"), data)
}
