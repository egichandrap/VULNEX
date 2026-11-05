package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrSetupTokenExpired     = errors.New("setup token has expired")
	ErrSetupTokenAlreadyUsed = errors.New("setup token has already been claimed")
	ErrInvalidSetupToken     = errors.New("invalid setup token provided")
)

// SetupTokenID is the identifier for a SetupToken aggregate.
type SetupTokenID struct {
	value string
}

func NewSetupTokenID(val string) (SetupTokenID, error) {
	if strings.TrimSpace(val) == "" {
		return SetupTokenID{}, errors.New("setup token ID cannot be empty")
	}
	return SetupTokenID{value: strings.TrimSpace(val)}, nil
}

func GenerateSetupTokenID() (SetupTokenID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return SetupTokenID{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	val := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return SetupTokenID{value: val}, nil
}

func (s SetupTokenID) String() string {
	return s.value
}

// SetupToken is the Aggregate Root governing initial administrator bootstrapping.
type SetupToken struct {
	id        SetupTokenID
	tokenHash string // Stored as SHA-256 hex string
	expiresAt time.Time
	usedAt    *time.Time
	createdAt time.Time
}

// GenerateSetupToken creates a new SetupToken and returns the aggregate and the raw high-entropy secret.
// The raw secret is displayed once in the terminal; only tokenHash is persisted.
func GenerateSetupToken(ttl time.Duration) (*SetupToken, string, error) {
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, "", fmt.Errorf("failed to read secure random bytes: %w", err)
	}
	rawSecret := hex.EncodeToString(rawBytes)

	hashBytes := sha256.Sum256([]byte(rawSecret))
	tokenHash := hex.EncodeToString(hashBytes[:])

	id, err := GenerateSetupTokenID()
	if err != nil {
		return nil, "", err
	}

	now := time.Now().UTC()
	return &SetupToken{
		id:        id,
		tokenHash: tokenHash,
		expiresAt: now.Add(ttl),
		usedAt:    nil,
		createdAt: now,
	}, rawSecret, nil
}

// ReconstituteSetupToken reconstitutes a token from the database.
func ReconstituteSetupToken(id SetupTokenID, tokenHash string, expiresAt time.Time, usedAt *time.Time, createdAt time.Time) *SetupToken {
	return &SetupToken{
		id:        id,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		usedAt:    usedAt,
		createdAt: createdAt,
	}
}

func (s *SetupToken) ID() SetupTokenID {
	return s.id
}

func (s *SetupToken) TokenHash() string {
	return s.tokenHash
}

func (s *SetupToken) ExpiresAt() time.Time {
	return s.expiresAt
}

func (s *SetupToken) UsedAt() *time.Time {
	return s.usedAt
}

func (s *SetupToken) CreatedAt() time.Time {
	return s.createdAt
}

func (s *SetupToken) IsClaimed() bool {
	return s.usedAt != nil
}

// Claim verifies the raw secret, checks expiration and previous usage, and marks the token as used.
func (s *SetupToken) Claim(rawSecret string, at time.Time) error {
	if s.usedAt != nil {
		return ErrSetupTokenAlreadyUsed
	}
	if at.After(s.expiresAt) {
		return ErrSetupTokenExpired
	}

	providedHashBytes := sha256.Sum256([]byte(rawSecret))
	providedHash := hex.EncodeToString(providedHashBytes[:])

	// Constant time comparison to prevent timing side channels
	if subtle.ConstantTimeCompare([]byte(s.tokenHash), []byte(providedHash)) != 1 {
		return ErrInvalidSetupToken
	}

	now := at.UTC()
	s.usedAt = &now
	return nil
}
