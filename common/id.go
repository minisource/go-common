package common

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
)

// NewID generates a new UUID
func NewID() uuid.UUID {
	return uuid.New()
}

// GenerateUniqueKey generates a cryptographically secure random key
func GenerateUniqueKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to UUID if random fails
		return uuid.New().String()
	}
	return hex.EncodeToString(bytes)
}

// GenerateToken generates a secure token of specified length
func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
