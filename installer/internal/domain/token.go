package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateEphemeralSetupToken creates a cryptographically secure 32-byte hex secret (ADR-013).
func GenerateEphemeralSetupToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
