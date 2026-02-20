package security

import (
	"golang.org/x/crypto/bcrypt"

	"go-auth/internal/domain"
)

type hasher struct {
	cost int
}

func NewHasher(cost int) domain.PasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	return &hasher{cost: cost}
}

func (h *hasher) Hash(password string) (domain.Password, error) {
	if password == "" {
		return domain.Password{}, domain.ErrPasswordRequired
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return domain.Password{}, err
	}

	return domain.NewPasswordFromHash(string(hash))
}

func (h *hasher) Compare(plainText string, hash domain.Password) bool {
	if plainText == "" {
		return false
	}

	if hash.IsZero() {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(hash.Hash()), []byte(plainText)) == nil
}
