package domain

import (
	"github.com/google/uuid"
)

type PasswordHasher interface {
	Hash(password string) (Password, error)
	Compare(plainText string, hash Password) bool
}

type OpaqueTokenManager interface {
	Generate() (string, error)
	Hash(token string) (string, error)
}

type AccessClaims struct {
	UserID uuid.UUID
	Role   Role
}

type AccessTokenManager interface {
	Generate(claims AccessClaims) (string, error)
	Validate(token string) (*AccessClaims, error)
}
