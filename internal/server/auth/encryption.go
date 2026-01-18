package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// Sentinel errors for encryption operations
var (
	ErrInvalidKeyLength   = errors.New("encryption key must be 32 bytes (256 bits)")
	ErrCiphertextTooShort = errors.New("ciphertext too short")
	ErrDecryptionFailed   = errors.New("decryption failed")
)

// TokenEncryptor provides methods to encrypt and decrypt tokens using AES-256-GCM
type TokenEncryptor interface {
	Encrypt(plaintext string) ([]byte, error)
	Decrypt(ciphertext []byte) (string, error)
}

type aesGcmEncryptor struct {
	aead cipher.AEAD
}

// NewTokenEncryptor creates a new TokenEncryptor with the provided base64-encoded 256-bit key
func NewTokenEncryptor(keyBase64 string) (TokenEncryptor, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	if len(keyBytes) != 32 {
		return nil, ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &aesGcmEncryptor{aead: aead}, nil
}

// Encrypt encrypts the plaintext using AES-256-GCM
// Returns ciphertext with nonce prepended
func (e *aesGcmEncryptor) Encrypt(plaintext string) ([]byte, error) {
	// Generate random nonce
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate (nonce is prepended to ciphertext)
	ciphertext := e.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

// Decrypt decrypts the ciphertext using AES-256-GCM
// Expects ciphertext with nonce prepended
func (e *aesGcmEncryptor) Decrypt(ciphertext []byte) (string, error) {
	nonceSize := e.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrCiphertextTooShort
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt and verify authentication tag
	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return string(plaintext), nil
}
