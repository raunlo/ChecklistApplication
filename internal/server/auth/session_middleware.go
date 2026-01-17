package auth

import (
	"context"
	"log"
	"net/http"
	"time"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "session_id"
	ginContextUserId  = "userId" // Context key for storing user ID
)

// SessionValidator is an interface to avoid import cycles with service package
type SessionValidator interface {
	ValidateSession(ctx context.Context, sessionId string) (*domain.Session, domain.Error)
	RefreshSessionActivity(ctx context.Context, sessionId string) domain.Error
}

// ExtractUserIdFromGinContext extracts the user ID from the Gin context
func ExtractUserIdFromGinContext(c *gin.Context) (string, bool) {
	userId, exists := c.Get(ginContextUserId)
	if !exists {
		return "", false
	}
	userIdStr, ok := userId.(string)
	return userIdStr, ok
}

// SessionAuthMiddleware validates session-based authentication
func SessionAuthMiddleware(authSessionService SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Enforce HTTPS in production
		if gin.Mode() == gin.ReleaseMode {
			proto := c.GetHeader("X-Forwarded-Proto")
			isHTTPS := proto == "https" || c.Request.TLS != nil
			if !isHTTPS {
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "HTTPS required",
					"message": "Please use HTTPS.",
				})
				c.Abort()
				return
			}
		}

		// Skip auth for OPTIONS (CORS preflight)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get session ID from cookie
		sessionId, err := c.Cookie(SessionCookieName)
		if err != nil || sessionId == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required",
				"message": "Please log in to access this resource.",
			})
			c.Abort()
			return
		}

		// Validate session
		session, domainErr := authSessionService.ValidateSession(c.Request.Context(), sessionId)
		if domainErr != nil {
			log.Printf("[SessionAuthMiddleware] Session validation error: %v", domainErr)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication failed",
				"message": "Session validation failed. Please log in again.",
			})
			c.Abort()
			return
		}

		if session == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Session expired",
				"message": "Your session has expired. Please log in again.",
			})
			c.Abort()
			return
		}

		// Store user ID in context (same key as GoogleAuthMiddleware for compatibility)
		c.Set(ginContextUserId, session.UserId)

		// Async update last activity (fire-and-forget to avoid blocking)
		// Use background context with timeout to avoid cancellation after response
		go func(sid string) {
			timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := authSessionService.RefreshSessionActivity(timeoutCtx, sid); err != nil {
				log.Printf("[SessionAuthMiddleware] Failed to update session activity: %v", err)
			}
		}(sessionId)

		c.Next()
	}
}
