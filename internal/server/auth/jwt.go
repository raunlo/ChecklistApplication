package auth

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const ginContextUserId = "userId"

// Simple in-memory rate limiter
var requestCounts = make(map[string]int)
var requestTimes = make(map[string]time.Time)
var mu sync.Mutex

// RateLimitMiddleware limits requests per IP to prevent brute-force attacks
func RateLimitMiddleware(requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		if lastTime, exists := requestTimes[ip]; exists && now.Sub(lastTime) > window {
			delete(requestCounts, ip)
			delete(requestTimes, ip)
		}

		count := requestCounts[ip]
		if count >= requests {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Please try again later.",
			})
			c.Abort()
			return
		}

		requestCounts[ip] = count + 1
		requestTimes[ip] = now
		mu.Unlock()

		c.Next()
	}
}

// GoogleAuthMiddleware validates Google ID tokens on each request
// Uses Google's official idtoken library for secure token validation
//
// Authentication method: Expects HTTP-only, Secure, and SameSite cookies
// - NOTE: This middleware assumes the 'user_token' cookie is set with the
//   HttpOnly, Secure, and SameSite attributes elsewhere (e.g., during login).
//   It does not itself enforce or set these attributes. Ensure cookies are
//   configured securely at creation time to prevent XSS and CSRF attacks.
// - Secure: Prevents XSS token theft
// - Automatic: Browser handles cookie transmission
// - SSE Compatible: Works with EventSource API
func GoogleAuthMiddleware(idtokenValidator IdtokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Enforce HTTPS in production
		if c.Request.TLS == nil && gin.Mode() == gin.ReleaseMode {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "HTTPS required",
				"message": "Please use HTTPS.",
			})
			c.Abort()
			return
		}

		// Skip authentication for OPTIONS preflight requests (CORS)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get token from user_token cookie (httpOnly)
		var idToken string

		cookie, err := c.Request.Cookie("user_token")
		if err != nil || cookie == nil || cookie.Value == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required",
				"message": "Please log in to access this resource.",
			})
			c.Abort()
			return
		}
		idToken = cookie.Value

		// Validate ID token with Google using official library
		// This automatically verifies: signature, issuer, audience, and expiration
		userInfo, err := idtokenValidator.ParseIdToken(c.Request.Context(), idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication failed",
				"message": "Invalid or expired token. Please log in again.",
			})
			c.Abort()
			return
		}

		// Store user info in context for use in handlers
		c.Set(ginContextUserId, userInfo.ID)

		c.Next()
	}
}

func ExtractUserIdFromGinContext(c *gin.Context) (string, bool) {
	userId, exists := c.Get(ginContextUserId)
	if !exists {
		return "", false
	}
	return userId.(string), true
}
