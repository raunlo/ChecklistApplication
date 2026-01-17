package domain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

var ClientIdContextKey = clientIdContextKey{}

type (
	userIdContextKey       struct{}
	hashedUserIdContextKey struct{}
	clientIdContextKey     = struct{}
)

var UserIdContextKey = userIdContextKey{}
var HashedUserIdContextKey = hashedUserIdContextKey{}

// GetUserIdFromContext returns the real user ID (for database operations)
func GetUserIdFromContext(ctx context.Context) (string, Error) {
	userIdValue := ctx.Value(UserIdContextKey)
	userId, ok := userIdValue.(string)
	if !ok || userId == "" {
		return "", NewError("User ID not found in context or invalid", 401)
	}
	return userId, nil
}

// GetHashedUserIdFromContext returns the hashed user ID (for logging)
// Returns "unknown" if not found to avoid logging errors
func GetHashedUserIdFromContext(ctx context.Context) string {
	hashedValue := ctx.Value(HashedUserIdContextKey)
	if hashed, ok := hashedValue.(string); ok && hashed != "" {
		return hashed
	}
	return "unknown"
}

// HashUserId hashes a user ID for privacy-preserving logging
// Returns first 16 characters of SHA-256 hash for brevity
func HashUserId(userId string) string {
	hash := sha256.Sum256([]byte(userId))
	return hex.EncodeToString(hash[:])[:16]
}

// AddUserIdToContext adds both real and hashed user ID to context
func AddUserIdToContext(ctx context.Context, userId string) context.Context {
	ctx = context.WithValue(ctx, UserIdContextKey, userId)
	ctx = context.WithValue(ctx, HashedUserIdContextKey, HashUserId(userId))
	return ctx
}
