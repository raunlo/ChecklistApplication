package service

import (
	"context"
	"log"
	"time"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/repository"
	"com.raunlo.checklist/internal/server/auth"
)

type TokenRefreshService struct {
	sessionRepo repository.ISessionRepository
	encryptor   auth.TokenEncryptor
	googleOAuth *auth.GoogleOAuthConfig
}

func NewTokenRefreshService(
	sessionRepo repository.ISessionRepository,
	encryptor auth.TokenEncryptor,
	googleOAuth *auth.GoogleOAuthConfig,
) *TokenRefreshService {
	return &TokenRefreshService{
		sessionRepo: sessionRepo,
		encryptor:   encryptor,
		googleOAuth: googleOAuth,
	}
}

// StartBackgroundRefresh runs every 30 minutes and refreshes tokens expiring in the next hour
func (s *TokenRefreshService) StartBackgroundRefresh(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	log.Println("[TokenRefreshService] Background token refresh started")

	for {
		select {
		case <-ticker.C:
			s.refreshExpiringTokens(ctx)
		case <-ctx.Done():
			log.Println("[TokenRefreshService] Background token refresh stopped")
			return
		}
	}
}

func (s *TokenRefreshService) refreshExpiringTokens(ctx context.Context) {
	// Find sessions with tokens expiring in the next hour
	expiringBefore := time.Now().Add(1 * time.Hour)
	sessions, err := s.sessionRepo.FindSessionsNeedingTokenRefresh(ctx, expiringBefore)
	if err != nil {
		log.Printf("[TokenRefreshService] Failed to find sessions needing refresh: %v", err)
		return
	}

	if len(sessions) == 0 {
		return
	}

	log.Printf("[TokenRefreshService] Found %d sessions needing token refresh", len(sessions))

	for _, session := range sessions {
		if err := s.refreshSessionTokens(ctx, &session); err != nil {
			log.Printf("[TokenRefreshService] Failed to refresh tokens for session %s: %v", session.SessionId, err)
			// Invalidate session if refresh fails
			s.sessionRepo.InvalidateSession(ctx, session.SessionId, "token_refresh_failed")
		} else {
			log.Printf("[TokenRefreshService] Successfully refreshed tokens for session %s", session.SessionId)
		}
	}
}

func (s *TokenRefreshService) refreshSessionTokens(ctx context.Context, session *domain.Session) error {
	// Decrypt refresh token
	refreshToken, err := s.encryptor.Decrypt(session.RefreshTokenEncrypted)
	if err != nil {
		return err
	}

	// Call Google OAuth token refresh endpoint
	newTokens, err := s.googleOAuth.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		return err
	}

	// Encrypt new access token
	newAccessTokenEnc, err := s.encryptor.Encrypt(newTokens.AccessToken)
	if err != nil {
		return err
	}

	// Update refresh token if Google provided a new one
	newRefreshTokenEnc := session.RefreshTokenEncrypted
	if newTokens.RefreshToken != "" {
		newRefreshTokenEnc, err = s.encryptor.Encrypt(newTokens.RefreshToken)
		if err != nil {
			return err
		}
	}

	// Calculate new expiration time
	newExpiresAt := time.Now().Add(time.Duration(newTokens.ExpiresIn) * time.Second)

	// Update session in database
	domainErr := s.sessionRepo.UpdateSessionTokens(
		ctx,
		session.SessionId,
		newAccessTokenEnc,
		newRefreshTokenEnc,
		newExpiresAt,
	)

	if domainErr != nil {
		return domainErr
	}

	return nil
}
