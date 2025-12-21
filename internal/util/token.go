package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateSecureToken generates a cryptographically secure random token
// using 48 random bytes which encodes to a 64-character base64 URL-safe string.
func GenerateSecureToken() (string, error) {
	bytes := make([]byte, 48) // 48 bytes = 64 base64 chars
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
