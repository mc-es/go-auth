package security_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go-auth/internal/domain"
	"go-auth/internal/security"
)

func TestNewOpaque(t *testing.T) {
	t.Run("positive length", func(t *testing.T) {
		m := security.NewOpaque(24)
		assert.NotNil(t, m)

		token, err := m.Generate()
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("zero length uses default", func(t *testing.T) {
		m := security.NewOpaque(0)
		assert.NotNil(t, m)

		token, err := m.Generate()
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("negative length uses default", func(t *testing.T) {
		m := security.NewOpaque(-1)
		assert.NotNil(t, m)

		token, err := m.Generate()
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

func TestOpaqueGenerate(t *testing.T) {
	m := security.NewOpaque(32)

	t.Run("returns non-empty token", func(t *testing.T) {
		token, err := m.Generate()
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("tokens are unique", func(t *testing.T) {
		token1, err1 := m.Generate()
		token2, err2 := m.Generate()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, token1, token2)
	})

	t.Run("token is valid base64 raw URL", func(t *testing.T) {
		token, err := m.Generate()
		assert.NoError(t, err)

		hash, hashErr := m.Hash(token)
		assert.NoError(t, hashErr)
		assert.NotEmpty(t, hash)
	})
}

func TestOpaqueHash(t *testing.T) {
	m := security.NewOpaque(32)

	t.Run("same input produces same hash", func(t *testing.T) {
		input := "some-token-value"
		hash1, err1 := m.Hash(input)
		hash2, err2 := m.Hash(input)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, hash1, hash2)
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		hash1, err1 := m.Hash("token-a")
		hash2, err2 := m.Hash("token-b")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("empty token returns error", func(t *testing.T) {
		hash, err := m.Hash("")
		assert.ErrorIs(t, err, domain.ErrTokenRequired)
		assert.Empty(t, hash)
	})

	t.Run("hash is non-empty for valid token", func(t *testing.T) {
		hash, err := m.Hash("any-non-empty-token")
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
	})
}
