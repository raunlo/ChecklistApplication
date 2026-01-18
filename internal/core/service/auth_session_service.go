package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"time"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/repository"
	"com.raunlo.checklist/internal/server/auth"
)

type IAuthSessionService interface {
	CreateSession(ctx context.Context, userId string, accessToken, refreshToken string, accessTokenExpiresAt time.Time) (string, domain.Error)
	ValidateSession(ctx context.Context, sessionId string) (*domain.Session, domain.Error)
	RefreshSessionActivity(ctx context.Context, sessionId string) domain.Error
	InvalidateSession(ctx context.Context, sessionId string, reason string) domain.Error
	GetDecryptedTokens(session *domain.Session) (accessToken, refreshToken string, err error)
	HandleOAuthCallback(ctx context.Context, code string) (sessionId string, domainErr domain.Error)
	// HandleDevLogin creates a dev user and session without OAuth (dev mode only)
	HandleDevLogin(ctx context.Context) (sessionId string, domainErr domain.Error)
	// RefreshTokensIfNeeded checks if Google access token is expired and refreshes it on-demand.
	// Also updates user info (name, photo) from Google when refreshing.
	// Returns nil if no refresh needed or refresh succeeded.
	RefreshTokensIfNeeded(ctx context.Context, session *domain.Session) domain.Error
}

type authSessionServiceImpl struct {
	sessionRepo  repository.ISessionRepository
	encryptor    auth.TokenEncryptor
	googleOAuth  *auth.GoogleOAuthConfig
	userService  IUserService
}

func NewAuthSessionService(
	sessionRepo repository.ISessionRepository,
	encryptor auth.TokenEncryptor,
	googleOAuth *auth.GoogleOAuthConfig,
	userService IUserService,
) IAuthSessionService {
	return &authSessionServiceImpl{
		sessionRepo: sessionRepo,
		encryptor:   encryptor,
		googleOAuth: googleOAuth,
		userService: userService,
	}
}

func (s *authSessionServiceImpl) CreateSession(ctx context.Context, userId string, accessToken, refreshToken string, accessTokenExpiresAt time.Time) (string, domain.Error) {
	// Generate cryptographically secure session ID
	sessionIdBytes := make([]byte, 32)
	if _, err := rand.Read(sessionIdBytes); err != nil {
		return "", domain.Wrap(err, "Failed to generate session ID", 500)
	}
	sessionId := base64.URLEncoding.EncodeToString(sessionIdBytes)

	// Encrypt tokens
	accessTokenEnc, err := s.encryptor.Encrypt(accessToken)
	if err != nil {
		return "", domain.Wrap(err, "Failed to encrypt access token", 500)
	}

	refreshTokenEnc, err := s.encryptor.Encrypt(refreshToken)
	if err != nil {
		return "", domain.Wrap(err, "Failed to encrypt refresh token", 500)
	}

	now := time.Now()
	session := domain.Session{
		SessionId:             sessionId,
		UserId:                userId,
		AccessTokenEncrypted:  accessTokenEnc,
		RefreshTokenEncrypted: refreshTokenEnc,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		CreatedAt:             now,
		LastActivityAt:        now,
		ExpiresAt:             now.Add(domain.MaxSessionLifetime),
	}

	_, domainErr := s.sessionRepo.CreateSession(ctx, session)
	if domainErr != nil {
		return "", domainErr
	}

	return sessionId, nil
}

func (s *authSessionServiceImpl) ValidateSession(ctx context.Context, sessionId string) (*domain.Session, domain.Error) {
	session, err := s.sessionRepo.FindSessionBySessionId(ctx, sessionId)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, nil
	}

	now := time.Now()

	// Check max session lifetime
	if now.After(session.ExpiresAt) {
		if err := s.sessionRepo.InvalidateSession(ctx, sessionId, "expired_max_lifetime"); err != nil {
			log.Printf("[WARN] Failed to invalidate expired session %s: %v", sessionId[:8], err)
		}
		return nil, nil
	}

	// Check idle timeout
	if now.Sub(session.LastActivityAt) > domain.IdleTimeout {
		if err := s.sessionRepo.InvalidateSession(ctx, sessionId, "expired_idle_timeout"); err != nil {
			log.Printf("[WARN] Failed to invalidate idle session %s: %v", sessionId[:8], err)
		}
		return nil, nil
	}

	return session, nil
}

func (s *authSessionServiceImpl) RefreshSessionActivity(ctx context.Context, sessionId string) domain.Error {
	return s.sessionRepo.UpdateSessionActivity(ctx, sessionId, time.Now())
}

func (s *authSessionServiceImpl) InvalidateSession(ctx context.Context, sessionId string, reason string) domain.Error {
	return s.sessionRepo.InvalidateSession(ctx, sessionId, reason)
}

func (s *authSessionServiceImpl) GetDecryptedTokens(session *domain.Session) (accessToken, refreshToken string, err error) {
	accessToken, err = s.encryptor.Decrypt(session.AccessTokenEncrypted)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.encryptor.Decrypt(session.RefreshTokenEncrypted)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// HandleOAuthCallback handles the complete OAuth callback flow
// Business logic: exchange code -> get user info -> create/update user -> create session
func (s *authSessionServiceImpl) HandleOAuthCallback(ctx context.Context, code string) (string, domain.Error) {
	// Exchange authorization code for tokens
	tokens, err := s.googleOAuth.ExchangeCode(ctx, code)
	if err != nil {
		return "", domain.Wrap(err, "Token exchange failed", 500)
	}

	// Get user info from Google
	userInfo, err := s.googleOAuth.GetUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		return "", domain.Wrap(err, "Failed to get user info", 500)
	}

	// Validate email is verified
	if !userInfo.VerifiedEmail {
		return "", domain.NewError("Email not verified", 403)
	}

	// Create or update user in database
	// Audit timestamps (created_at, updated_at) are handled by SQL
	user := domain.User{
		UserId:   userInfo.ID,
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		PhotoUrl: userInfo.Picture,
	}

	if domainErr := s.userService.CreateOrUpdateUser(ctx, user); domainErr != nil {
		return "", domain.Wrap(domainErr, "Failed to create/update user", 500)
	}

	// Create session
	accessTokenExpiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	sessionId, domainErr := s.CreateSession(
		ctx,
		userInfo.ID,
		tokens.AccessToken,
		tokens.RefreshToken,
		accessTokenExpiresAt,
	)
	if domainErr != nil {
		return "", domainErr
	}

	return sessionId, nil
}

// HandleDevLogin creates a dev user and session without OAuth
// This should only be used in development mode
func (s *authSessionServiceImpl) HandleDevLogin(ctx context.Context) (string, domain.Error) {
	// Create or update dev user
	devUser := domain.User{
		UserId:   "dev-user-123",
		Email:    "dev@localhost",
		Name:     "Dev User",
		PhotoUrl: "",
	}

	if domainErr := s.userService.CreateOrUpdateUser(ctx, devUser); domainErr != nil {
		return "", domain.Wrap(domainErr, "Failed to create/update dev user", 500)
	}

	// Create session with dummy tokens (they won't be used in dev mode)
	accessTokenExpiresAt := time.Now().Add(24 * time.Hour)
	sessionId, domainErr := s.CreateSession(
		ctx,
		devUser.UserId,
		"dev-access-token",
		"dev-refresh-token",
		accessTokenExpiresAt,
	)
	if domainErr != nil {
		return "", domainErr
	}

	return sessionId, nil
}

// RefreshTokensIfNeeded checks if the Google access token is expired and refreshes it on-demand.
// This is designed for serverless environments (like Cloud Run) where background refresh isn't possible.
// When refreshing, it also fetches fresh user info (name, photo) from Google.
func (s *authSessionServiceImpl) RefreshTokensIfNeeded(ctx context.Context, session *domain.Session) domain.Error {
	// Add 5-minute buffer to refresh tokens slightly before they expire
	refreshBuffer := 5 * time.Minute
	if time.Now().Before(session.AccessTokenExpiresAt.Add(-refreshBuffer)) {
		// Token is still valid, no refresh needed
		return nil
	}

	log.Printf("[TokenRefresh] Access token expired for session %s, refreshing...", session.SessionId[:8])

	// Decrypt the refresh token
	_, refreshToken, err := s.GetDecryptedTokens(session)
	if err != nil {
		log.Printf("[TokenRefresh] Failed to decrypt refresh token: %v", err)
		_ = s.sessionRepo.InvalidateSession(ctx, session.SessionId, "token_decrypt_failed")
		return domain.Wrap(err, "Failed to decrypt refresh token", 500)
	}

	// Call Google to get new access token
	tokens, err := s.googleOAuth.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		log.Printf("[TokenRefresh] Failed to refresh access token from Google: %v", err)
		_ = s.sessionRepo.InvalidateSession(ctx, session.SessionId, "token_refresh_failed")
		return domain.Wrap(err, "Failed to refresh access token", 401)
	}

	// Fetch fresh user info from Google
	userInfo, err := s.googleOAuth.GetUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		log.Printf("[TokenRefresh] Failed to get user info from Google: %v", err)
		// Don't fail the request, just log it - the tokens are still valid
	} else {
		// Update user info in database
		user := domain.User{
			UserId:   session.UserId,
			Email:    userInfo.Email,
			Name:     userInfo.Name,
			PhotoUrl: userInfo.Picture,
		}
		if domainErr := s.userService.CreateOrUpdateUser(ctx, user); domainErr != nil {
			log.Printf("[TokenRefresh] Failed to update user info: %v", domainErr)
			// Don't fail the request, just log it
		}
	}

	// Encrypt new tokens
	accessTokenEnc, err := s.encryptor.Encrypt(tokens.AccessToken)
	if err != nil {
		return domain.Wrap(err, "Failed to encrypt new access token", 500)
	}

	// Use new refresh token if provided, otherwise keep the old one
	var refreshTokenEnc []byte
	if tokens.RefreshToken != "" {
		refreshTokenEnc, err = s.encryptor.Encrypt(tokens.RefreshToken)
		if err != nil {
			return domain.Wrap(err, "Failed to encrypt new refresh token", 500)
		}
	} else {
		// Google didn't provide a new refresh token, keep using the old one
		refreshTokenEnc = session.RefreshTokenEncrypted
	}

	// Calculate new expiration time
	accessTokenExpiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)

	// Update session in database
	if domainErr := s.sessionRepo.UpdateSessionTokens(ctx, session.SessionId, accessTokenEnc, refreshTokenEnc, accessTokenExpiresAt); domainErr != nil {
		return domainErr
	}

	log.Printf("[TokenRefresh] Successfully refreshed tokens for session %s", session.SessionId[:8])
	return nil
}
