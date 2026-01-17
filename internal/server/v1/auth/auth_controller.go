package auth

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/service"
	serverAuth "com.raunlo.checklist/internal/server/auth"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	googleOAuth    *serverAuth.GoogleOAuthConfig
	authSessionSvc service.IAuthSessionService
	userService    service.IUserService
	frontendUrl    serverAuth.FrontendUrl
}

func NewAuthController(
	googleOAuth *serverAuth.GoogleOAuthConfig,
	authSessionSvc service.IAuthSessionService,
	userService service.IUserService,
	frontendUrl serverAuth.FrontendUrl,
) *AuthController {
	return &AuthController{
		googleOAuth:    googleOAuth,
		authSessionSvc: authSessionSvc,
		userService:    userService,
		frontendUrl:    frontendUrl,
	}
}

// InitiateGoogleLogin redirects to Google OAuth
// GET /api/v1/auth/google/login
func (ctrl *AuthController) InitiateGoogleLogin(c *gin.Context) {
	// Generate state token for CSRF protection
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state token"})
		return
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Store state in httpOnly cookie (short-lived, 5 minutes)
	// Use SameSite=Lax for OAuth flow (Google redirects back to our domain)
	isProduction := gin.Mode() == gin.ReleaseMode
	stateCookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   300, // 5 minutes
		Path:     "/",
		Domain:   "",                   // Empty = host-only cookie (more reliable for OAuth)
		Secure:   isProduction,         // HTTPS only in production
		HttpOnly: true,                 // Not accessible via JavaScript
		SameSite: http.SameSiteLaxMode, // Allow OAuth redirect
	}
	http.SetCookie(c.Writer, stateCookie)

	// Redirect to Google OAuth
	authURL := ctrl.googleOAuth.GetAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// HandleGoogleCallback processes OAuth callback from Google
// GET /api/v1/auth/google/callback
func (ctrl *AuthController) HandleGoogleCallback(c *gin.Context) {
	// Validate state token (CSRF protection)
	state := c.Query("state")
	cookieState, err := c.Cookie("oauth_state")

	if err != nil || state != cookieState || state == "" {
		log.Printf("[AuthController] State validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state token"})
		return
	}

	// Clear state cookie
	isProduction := gin.Mode() == gin.ReleaseMode
	clearStateCookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "",
		Secure:   isProduction,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(c.Writer, clearStateCookie)

	// Handle OAuth error
	if errorParam := c.Query("error"); errorParam != "" {
		log.Printf("[AuthController] OAuth error: %s", errorParam)
		c.JSON(http.StatusBadRequest, gin.H{"error": errorParam})
		return
	}

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	// Handle OAuth callback business logic (delegated to service)
	sessionId, domainErr := ctrl.authSessionSvc.HandleOAuthCallback(c.Request.Context(), code)
	if domainErr != nil {
		log.Printf("[AuthController] OAuth callback failed: %v", domainErr)
		c.JSON(domainErr.ResponseCode(), gin.H{"error": domainErr.Error()})
		return
	}

	// Set session cookie with SameSite=Lax
	sessionCookie := &http.Cookie{
		Name:     serverAuth.SessionCookieName,
		Value:    sessionId,
		MaxAge:   int(domain.MaxSessionLifetime.Seconds()),
		Path:     "/",
		Domain:   "",
		Secure:   isProduction,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, sessionCookie)

	// Redirect to frontend checklist page directly
	c.Redirect(http.StatusTemporaryRedirect, string(ctrl.frontendUrl)+"/checklist")
}

// Logout invalidates the session
// POST /api/v1/auth/logout
func (ctrl *AuthController) Logout(c *gin.Context) {
	sessionId, err := c.Cookie(serverAuth.SessionCookieName)
	if err == nil && sessionId != "" {
		if domainErr := ctrl.authSessionSvc.InvalidateSession(c.Request.Context(), sessionId, "logout"); domainErr != nil {
			log.Printf("[AuthController] Failed to invalidate session: %v", domainErr)
		}
	}

	// Clear session cookie
	isProduction := gin.Mode() == gin.ReleaseMode
	clearCookie := &http.Cookie{
		Name:     serverAuth.SessionCookieName,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "",
		Secure:   isProduction,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, clearCookie)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetSession returns the current session status with user details
// GET /api/v1/auth/session
func (ctrl *AuthController) GetSession(c *gin.Context) {
	sessionId, err := c.Cookie(serverAuth.SessionCookieName)
	if err != nil || sessionId == "" {
		c.JSON(http.StatusOK, gin.H{"authenticated": false})
		return
	}

	session, domainErr := ctrl.authSessionSvc.ValidateSession(c.Request.Context(), sessionId)
	if domainErr != nil || session == nil {
		c.JSON(http.StatusOK, gin.H{"authenticated": false})
		return
	}

	// Fetch user details
	response := gin.H{
		"authenticated": true,
		"userId":        session.UserId,
	}

	user, userErr := ctrl.userService.GetUserById(c.Request.Context(), session.UserId)
	if userErr == nil && user != nil {
		response["user"] = gin.H{
			"name":     user.Name,
			"photoUrl": user.PhotoUrl,
			"email":    user.Email,
		}
	}

	c.JSON(http.StatusOK, response)
}

// DevLogin creates a dev session without OAuth (dev mode only)
// GET /api/v1/auth/dev/login
func (ctrl *AuthController) DevLogin(c *gin.Context) {
	// Only allow in non-production mode
	if gin.Mode() == gin.ReleaseMode {
		c.JSON(http.StatusForbidden, gin.H{"error": "Dev login not available in production"})
		return
	}

	sessionId, domainErr := ctrl.authSessionSvc.HandleDevLogin(c.Request.Context())
	if domainErr != nil {
		log.Printf("[AuthController] Dev login failed: %v", domainErr)
		c.JSON(domainErr.ResponseCode(), gin.H{"error": domainErr.Error()})
		return
	}

	// Set session cookie
	sessionCookie := &http.Cookie{
		Name:     serverAuth.SessionCookieName,
		Value:    sessionId,
		MaxAge:   int(domain.MaxSessionLifetime.Seconds()),
		Path:     "/",
		Domain:   "",
		Secure:   false, // Dev mode uses HTTP
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, sessionCookie)

	// Redirect to frontend checklist page directly
	c.Redirect(http.StatusTemporaryRedirect, string(ctrl.frontendUrl)+"/checklist")
}
