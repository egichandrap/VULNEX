package crypto

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordHasher implements domain.PasswordHasher using bcrypt.
type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher(cost int) *BcryptPasswordHasher {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptPasswordHasher{cost: cost}
}

func (h *BcryptPasswordHasher) Hash(plainPassword string) (string, error) {
	if plainPassword == "" {
		return "", errors.New("cannot hash empty password")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), h.cost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hashing failed: %w", err)
	}
	return string(bytes), nil
}

func (h *BcryptPasswordHasher) Compare(plainPassword, passwordHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(plainPassword))
	return err == nil
}
