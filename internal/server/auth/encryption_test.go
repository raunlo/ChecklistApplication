package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenEncryptor(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr error
	}{
		{
			name:    "valid 32-byte key",
			key:     base64.StdEncoding.EncodeToString(make([]byte, 32)),
			wantErr: nil,
		},
		{
			name:    "invalid key length - 16 bytes",
			key:     base64.StdEncoding.EncodeToString(make([]byte, 16)),
			wantErr: ErrInvalidKeyLength,
		},
		{
			name:    "invalid key length - 64 bytes",
			key:     base64.StdEncoding.EncodeToString(make([]byte, 64)),
			wantErr: ErrInvalidKeyLength,
		},
		{
			name:    "invalid base64",
			key:     "not-valid-base64!!!",
			wantErr: nil, // Wrapped error, check it exists
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: nil, // Wrapped error for base64 decode
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := NewTokenEncryptor(tt.key)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, ErrInvalidKeyLength) {
					assert.ErrorIs(t, err, ErrInvalidKeyLength)
				}
				assert.Nil(t, enc)
			} else if tt.key == "" || tt.key == "not-valid-base64!!!" {
				// These should fail with wrapped errors
				require.Error(t, err)
				assert.Nil(t, enc)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, enc)
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	// Generate valid key
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	require.NoError(t, err)
	key := base64.StdEncoding.EncodeToString(keyBytes)

	enc, err := NewTokenEncryptor(key)
	require.NoError(t, err)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"short string", "hello"},
		{"medium string", "this is a test token"},
		{"long string", "this is a very long token that should be encrypted properly with AES-256-GCM encryption"},
		{"special chars", "!@#$%^&*()_+-={}[]|\\:\";<>?,./"},
		{"unicode", "こんにちは世界🌍🚀"},
		{"json-like", `{"user":"john","token":"abc123"}`},
		{"base64-like", "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo="},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := enc.Encrypt(tt.plaintext)
			require.NoError(t, err)
			assert.NotEmpty(t, ciphertext)

			// Ciphertext should be different from plaintext
			if tt.plaintext != "" {
				assert.NotEqual(t, []byte(tt.plaintext), ciphertext)
			}

			// Ciphertext should include nonce (12 bytes for GCM) + encrypted data + auth tag (16 bytes)
			// For empty string, it's exactly 12 + 16 = 28 bytes
			assert.GreaterOrEqual(t, len(ciphertext), 12+16)

			// Decrypt
			decrypted, err := enc.Decrypt(ciphertext)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestEncrypt_DifferentCiphertext(t *testing.T) {
	// Generate valid key
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	require.NoError(t, err)
	key := base64.StdEncoding.EncodeToString(keyBytes)

	enc, err := NewTokenEncryptor(key)
	require.NoError(t, err)

	plaintext := "same plaintext"

	// Encrypt same plaintext twice
	ciphertext1, err := enc.Encrypt(plaintext)
	require.NoError(t, err)

	ciphertext2, err := enc.Encrypt(plaintext)
	require.NoError(t, err)

	// Ciphertexts should be different (different nonces)
	assert.NotEqual(t, ciphertext1, ciphertext2)

	// But both should decrypt to the same plaintext
	decrypted1, err := enc.Decrypt(ciphertext1)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted1)

	decrypted2, err := enc.Decrypt(ciphertext2)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted2)
}

func TestDecrypt_Errors(t *testing.T) {
	// Generate valid key
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	require.NoError(t, err)
	key := base64.StdEncoding.EncodeToString(keyBytes)

	enc, err := NewTokenEncryptor(key)
	require.NoError(t, err)

	tests := []struct {
		name       string
		ciphertext []byte
		wantErr    error
	}{
		{
			name:       "ciphertext too short - empty",
			ciphertext: []byte{},
			wantErr:    ErrCiphertextTooShort,
		},
		{
			name:       "ciphertext too short - 5 bytes",
			ciphertext: []byte{1, 2, 3, 4, 5},
			wantErr:    ErrCiphertextTooShort,
		},
		{
			name:       "ciphertext too short - 11 bytes (nonce is 12)",
			ciphertext: make([]byte, 11),
			wantErr:    ErrCiphertextTooShort,
		},
		{
			name:       "tampered ciphertext - random bytes",
			ciphertext: make([]byte, 50), // 12 + 38 bytes
			wantErr:    ErrDecryptionFailed,
		},
		{
			name:       "nil ciphertext",
			ciphertext: nil,
			wantErr:    ErrCiphertextTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := enc.Decrypt(tt.ciphertext)
			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestDecrypt_TamperedCiphertext(t *testing.T) {
	// Generate valid key
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	require.NoError(t, err)
	key := base64.StdEncoding.EncodeToString(keyBytes)

	enc, err := NewTokenEncryptor(key)
	require.NoError(t, err)

	// Encrypt valid plaintext
	plaintext := "secure token"
	ciphertext, err := enc.Encrypt(plaintext)
	require.NoError(t, err)

	// Tamper with ciphertext (flip a bit in the middle)
	tamperedCiphertext := make([]byte, len(ciphertext))
	copy(tamperedCiphertext, ciphertext)
	tamperedCiphertext[len(tamperedCiphertext)/2] ^= 0xFF

	// Decryption should fail (GCM authentication)
	_, err = enc.Decrypt(tamperedCiphertext)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDecryptionFailed)
}

func TestDecrypt_DifferentKey(t *testing.T) {
	// Generate two different keys
	keyBytes1 := make([]byte, 32)
	_, err := rand.Read(keyBytes1)
	require.NoError(t, err)
	key1 := base64.StdEncoding.EncodeToString(keyBytes1)

	keyBytes2 := make([]byte, 32)
	_, err = rand.Read(keyBytes2)
	require.NoError(t, err)
	key2 := base64.StdEncoding.EncodeToString(keyBytes2)

	enc1, err := NewTokenEncryptor(key1)
	require.NoError(t, err)

	enc2, err := NewTokenEncryptor(key2)
	require.NoError(t, err)

	// Encrypt with first key
	plaintext := "secret data"
	ciphertext, err := enc1.Encrypt(plaintext)
	require.NoError(t, err)

	// Decrypt with second key should fail
	_, err = enc2.Decrypt(ciphertext)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDecryptionFailed)
}

// Benchmark encryption
func BenchmarkEncrypt(b *testing.B) {
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	require.NoError(b, err)
	key := base64.StdEncoding.EncodeToString(keyBytes)

	enc, err := NewTokenEncryptor(key)
	require.NoError(b, err)

	plaintext := "this is a test token that needs to be encrypted"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := enc.Encrypt(plaintext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark decryption
func BenchmarkDecrypt(b *testing.B) {
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	require.NoError(b, err)
	key := base64.StdEncoding.EncodeToString(keyBytes)

	enc, err := NewTokenEncryptor(key)
	require.NoError(b, err)

	plaintext := "this is a test token that needs to be encrypted"
	ciphertext, err := enc.Encrypt(plaintext)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := enc.Decrypt(ciphertext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark parallel encryption
func BenchmarkEncrypt_Parallel(b *testing.B) {
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	require.NoError(b, err)
	key := base64.StdEncoding.EncodeToString(keyBytes)

	enc, err := NewTokenEncryptor(key)
	require.NoError(b, err)

	plaintext := "this is a test token that needs to be encrypted"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := enc.Encrypt(plaintext)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
