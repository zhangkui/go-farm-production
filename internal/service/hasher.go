package service

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// bcryptHasher is the production Hasher using bcrypt with a configurable cost.
type bcryptHasher struct {
	cost int
}

// NewHasher builds a Hasher with the given bcrypt cost factor.
func NewHasher(cost int) Hasher { return &bcryptHasher{cost: cost} }

func (h *bcryptHasher) Hash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(b), nil
}

func (h *bcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
