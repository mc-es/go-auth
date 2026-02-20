package domain

import (
	"github.com/google/uuid"
)

type PasswordHasher interface {
	Hash(password string) (Password, error)
	Compare(plainText string, hash Password) bool
}

type Claims struct {
	UserID uuid.UUID
	Role   Role
	Type   TokenType
}

type TokenManager interface {
	Generate(claims Claims) (string, error)
	Validate(token string) (*Claims, error)
}
