package util

import (
	"encoding/base64"
	"testing"
)

func TestGenerateSecureToken(t *testing.T) {
	t.Run("generates valid base64 URL-safe token", func(t *testing.T) {
		token, err := GenerateSecureToken()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(token) != 64 {
			t.Errorf("expected token length 64, got %d", len(token))
		}

		// Verify it's valid base64 URL-safe encoding
		decoded, err := base64.URLEncoding.DecodeString(token)
		if err != nil {
			t.Errorf("token is not valid base64 URL-safe: %v", err)
		}

		if len(decoded) != 48 {
			t.Errorf("expected 48 decoded bytes, got %d", len(decoded))
		}
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		tokens := make(map[string]bool)
		iterations := 1000

		for i := 0; i < iterations; i++ {
			token, err := GenerateSecureToken()
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}

			if tokens[token] {
				t.Errorf("duplicate token generated: %s", token)
			}
			tokens[token] = true
		}

		if len(tokens) != iterations {
			t.Errorf("expected %d unique tokens, got %d", iterations, len(tokens))
		}
	})

	t.Run("tokens contain no special characters requiring escaping", func(t *testing.T) {
		token, err := GenerateSecureToken()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Base64 URL-safe should only contain: A-Z, a-z, 0-9, -, _
		for _, char := range token {
			if !((char >= 'A' && char <= 'Z') ||
				(char >= 'a' && char <= 'z') ||
				(char >= '0' && char <= '9') ||
				char == '-' || char == '_') {
				t.Errorf("token contains invalid character: %c", char)
			}
		}
	})
}
