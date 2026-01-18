package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CSRFCookieName  = "csrf_token"
	CSRFHeaderName  = "X-CSRF-Token"
	CSRFTokenLength = 32
)

// GenerateCSRFToken creates a cryptographically secure random token
func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SetCSRFTokenMiddleware generates and sets CSRF token if not present
// The token is set in both a cookie (HttpOnly) and response header (for frontend to read)
func SetCSRFTokenMiddleware(isProduction bool, csrfCookieDomain string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if CSRF token cookie already exists
		if _, err := c.Cookie(CSRFCookieName); err != nil {
			// Generate new token
			token, err := GenerateCSRFToken()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Token generation failed",
					"message": "Failed to generate CSRF token",
				})
				c.Abort()
				return
			}

			// Set cookie - NOT HttpOnly so JavaScript can read it for header
			// CSRF tokens don't need to be secret (unlike auth tokens)
			// Protection comes from SameSite and same-origin policy
			sameSite := http.SameSiteStrictMode
			if !isProduction {
				// In dev mode, use Lax to allow cross-subdomain requests
				sameSite = http.SameSiteLaxMode
			}
			cookie := &http.Cookie{
				Name:     CSRFCookieName,
				Value:    token,
				MaxAge:   60 * 60 * 24 * 7, // 7 days
				Domain:   csrfCookieDomain,
				Path:     "/",
				Secure:   isProduction, // HTTPS only in production
				HttpOnly: false,        // JavaScript must read this
				SameSite: sameSite,
			}
			http.SetCookie(c.Writer, cookie)

			// Also set in response header for immediate access
			c.Header(CSRFHeaderName, token)
		}

		c.Next()
	}
}

// CSRFMiddleware validates CSRF tokens on state-changing methods
// Uses double-submit cookie pattern: validates that cookie token matches header token
func CSRFMiddleware(isProduction bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		// Skip validation for safe methods and OPTIONS (CORS preflight)
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		// Get token from cookie
		cookieToken, err := c.Cookie(CSRFCookieName)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "CSRF token missing",
				"message": "CSRF token cookie not found. Please refresh the page.",
			})
			c.Abort()
			return
		}

		// Get token from header
		headerToken := c.GetHeader(CSRFHeaderName)
		if headerToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "CSRF token missing",
				"message": "X-CSRF-Token header required for state-changing requests",
			})
			c.Abort()
			return
		}

		// Validate tokens match using constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "CSRF token invalid",
				"message": "CSRF token mismatch. This request may be forged.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
