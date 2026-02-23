package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"go-auth/internal/domain"
)

type opaqueManager struct {
	length int
}

func NewOpaque(length int) domain.OpaqueTokenManager {
	if length <= 0 {
		length = 32
	}

	return &opaqueManager{length: length}
}

func (m *opaqueManager) Generate() (string, error) {
	b := make([]byte, m.length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (m *opaqueManager) Hash(token string) (string, error) {
	if token == "" {
		return "", domain.ErrTokenRequired
	}

	sum := sha256.Sum256([]byte(token))

	return base64.RawURLEncoding.EncodeToString(sum[:]), nil
}
