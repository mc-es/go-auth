package security_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"go-auth/internal/domain"
	"go-auth/internal/security"
)

func TestNewHasher(t *testing.T) {
	t.Run("cost in range", func(t *testing.T) {
		h := security.NewHasher(bcrypt.MinCost)
		assert.NotNil(t, h)

		hash, err := h.Hash("password")
		assert.NoError(t, err)
		assert.False(t, hash.IsZero())
		assert.True(t, h.Compare("password", hash))
	})

	t.Run("cost below min uses default", func(t *testing.T) {
		h := security.NewHasher(0)
		assert.NotNil(t, h)

		hash, err := h.Hash("password")
		assert.NoError(t, err)
		assert.False(t, hash.IsZero())
		assert.True(t, h.Compare("password", hash))
	})

	t.Run("cost above max uses default", func(t *testing.T) {
		h := security.NewHasher(100)
		assert.NotNil(t, h)

		hash, err := h.Hash("password")
		assert.NoError(t, err)
		assert.False(t, hash.IsZero())
		assert.True(t, h.Compare("password", hash))
	})
}

func TestHasherHash(t *testing.T) {
	h := security.NewHasher(bcrypt.MinCost)

	t.Run("valid password", func(t *testing.T) {
		hash, err := h.Hash("secret")
		assert.NoError(t, err)
		assert.False(t, hash.IsZero())
		assert.NotEmpty(t, hash.String())
	})

	t.Run("unique salts per hash", func(t *testing.T) {
		hash1, err1 := h.Hash("secret")
		hash2, err2 := h.Hash("secret")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, hash1.Hash(), hash2.Hash())
	})

	t.Run("empty password", func(t *testing.T) {
		hash, err := h.Hash("")
		assert.ErrorIs(t, err, domain.ErrPasswordRequired)
		assert.True(t, hash.IsZero())
	})
}

func TestHasherCompare(t *testing.T) {
	h := security.NewHasher(bcrypt.MinCost)
	hash, err := h.Hash("correct")
	assert.NoError(t, err)

	t.Run("match", func(t *testing.T) {
		assert.True(t, h.Compare("correct", hash))
	})

	t.Run("mismatch", func(t *testing.T) {
		assert.False(t, h.Compare("wrong", hash))
	})

	t.Run("empty plaintext rejected", func(t *testing.T) {
		assert.False(t, h.Compare("", hash))
	})

	t.Run("zero hash always fails", func(t *testing.T) {
		var zero domain.Password
		assert.False(t, h.Compare("", zero))
		assert.False(t, h.Compare("anything", zero))
	})
}
