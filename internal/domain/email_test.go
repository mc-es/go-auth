package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go-auth/internal/domain"
)

const (
	emailAlice = "alice@example.com"
	emailBob   = "bob@example.com"
	emailUser  = "user@example.com"
)

func TestEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "Valid simple email",
			input: "Bob@example.com",
			want:  emailBob,
		},
		{
			name:  "Valid email with uppercase and spaces",
			input: "  Alice@Example.COM  ",
			want:  emailAlice,
		},
		{
			name:  "Valid email with name tag",
			input: "John Doe <john@example.com>",
			want:  "john@example.com",
		},
		{
			name:    "Invalid empty",
			input:   "",
			wantErr: domain.ErrEmailRequired,
		},
		{
			name:    "Invalid format no @",
			input:   "testexample.com",
			wantErr: domain.ErrEmailInvalid,
		},
		{
			name:    "Invalid format no domain",
			input:   "test@",
			wantErr: domain.ErrEmailInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := domain.NewEmail(tt.input)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestEmailIsZero(t *testing.T) {
	var zero domain.Email
	assert.True(t, zero.IsZero())

	email, _ := domain.NewEmail(emailUser)
	assert.False(t, email.IsZero())
}

func TestEmailLocal(t *testing.T) {
	var zero domain.Email
	assert.Equal(t, "", zero.Local())

	email, _ := domain.NewEmail(emailAlice)
	assert.Equal(t, "alice", email.Local())
}

func TestEmailDomain(t *testing.T) {
	var zero domain.Email
	assert.Equal(t, "", zero.Domain())

	email, _ := domain.NewEmail(emailAlice)
	assert.Equal(t, "example.com", email.Domain())
}

func TestEmailMask(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"zero email", "", ""},
		{"short local (1 char)", "a@x.com", "a***@x.com"},
		{"short local (2 chars)", "ab@x.com", "ab***@x.com"},
		{"normal local", "alice@example.com", "a***e@example.com"},
		{"long local", "longname@example.com", "l***e@example.com"},
		{"unicode local", "çğş@example.com", "ç***ş@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			email, _ := domain.NewEmail(tt.email)
			assert.Equal(t, tt.want, email.Mask())
		})
	}
}

func TestEmailValue(t *testing.T) {
	var zero domain.Email

	val, err := zero.Value()
	assert.Empty(t, val)
	assert.ErrorIs(t, err, domain.ErrEmailRequired)

	email, _ := domain.NewEmail(emailUser)
	val, err = email.Value()
	assert.NoError(t, err)
	assert.NotEmpty(t, val)
	assert.Equal(t, emailUser, val)
}

func TestEmailScan(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr error
	}{
		{"nil", nil, "", nil},
		{"string", emailAlice, emailAlice, nil},
		{"bytes", []byte(emailBob), emailBob, nil},
		{"invalid type", 123, "", domain.ErrEmailScan},
		{"invalid email string", "not-an-email", "", domain.ErrEmailInvalid},
		{"empty string", "", "", domain.ErrEmailRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var e domain.Email

			err := e.Scan(tt.input)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, e.String())
		})
	}
}

func TestEmailMarshalText(t *testing.T) {
	var zero domain.Email

	data, err := zero.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte(""), data)

	email, _ := domain.NewEmail(emailUser)
	data, err = email.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, []byte(emailUser), data)
}

func TestEmailUnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    string
		wantErr error
	}{
		{"valid", []byte(emailAlice), emailAlice, nil},
		{"empty", []byte(""), "", nil},
		{"invalid", []byte("bad"), "", domain.ErrEmailInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var e domain.Email

			err := e.UnmarshalText(tt.input)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, e.String())
		})
	}
}
