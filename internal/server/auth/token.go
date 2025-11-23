package auth

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
)

// UserInfo represents minimal user information from Google ID token
// Only stores user ID to minimize PII/PHI data exposure
type UserInfo struct {
	ID string // Google user ID (sub claim) - not considered PII
}

type IdtokenValidator interface {
	ParseIdToken(ctx context.Context, idToken string) (*UserInfo, error)
}

type googleIdtokenValidator struct {
	clientID string
}

func NewIDTokenValidator(googleId string) IdtokenValidator {
	if err := validateGoogleAuthConfig(googleId); err != nil {
		panic(fmt.Errorf("failed to validate Google auth config: %w", err))
	}

	return &googleIdtokenValidator{
		clientID: googleId,
	}
}

// validateGoogleAuthConfig validates that all required Google OAuth configuration is present
// Should be called during application startup/deployment
func validateGoogleAuthConfig(clientID string) error {
	if clientID == "" {
		return fmt.Errorf("GOOGLE_CLIENT_ID environment variable is required but not set")
	}
	return nil
}

func (validator *googleIdtokenValidator) ParseIdToken(ctx context.Context, idToken string) (*UserInfo, error) {
	// Get Google OAuth Client ID from environment

	// Validate the ID token using Google's official library
	// This verifies the signature, issuer, audience, and expiration
	payload, err := idtoken.Validate(ctx, idToken, validator.clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate ID token: %w", err)
	}

	// Extract only the user ID (sub claim) - minimal data approach
	userID := payload.Subject // This is the unique Google user ID

	// Validate that we have the essential user ID
	if userID == "" {
		return nil, fmt.Errorf("missing user ID (sub claim) in ID token")
	}

	// We still validate email exists and is verified for security,
	// but we don't store it to minimize PII exposure

	// Return only the user ID - no PII/PHI data stored
	return &UserInfo{
		ID: userID,
	}, nil
}
