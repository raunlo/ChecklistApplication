package middleware

import (
	"log"
	"net/http"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/service"
	"github.com/gin-gonic/gin"
)

// ClientIdValidatorMiddleware validates server-generated client IDs
// If enforceValidation is false, it only logs warnings for invalid/missing IDs (migration mode)
// If enforceValidation is true, it rejects requests with invalid/missing IDs
func ClientIdValidatorMiddleware(sessionService service.ISessionService, enforceValidation bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract client ID from header or query parameter
		clientId := c.GetHeader("X-Client-Id")
		if clientId == "" {
			clientId = c.Query("clientId")
		}

		// Get user ID from context (set by auth middleware)
		userId, err := domain.GetUserIdFromContext(c.Request.Context())
		if err != nil {
			// If no user ID, auth middleware should have already rejected
			c.Next()
			return
		}

		// If no client ID provided
		if clientId == "" {
			if enforceValidation {
				log.Printf("Client ID missing for user(id=%s), path=%s", domain.HashUserId(userId), c.Request.URL.Path)
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Client ID required",
					"message": "X-Client-Id header or clientId query parameter is required. Call POST /api/v1/session/client-id to generate one.",
				})
				c.Abort()
				return
			} else {
				// Migration mode: just log warning
				log.Printf("WARNING: Client ID missing for user(id=%s), path=%s (migration mode)", domain.HashUserId(userId), c.Request.URL.Path)
				c.Next()
				return
			}
		}

		// Validate client ID
		valid := sessionService.ValidateClientId(clientId, userId)
		if !valid {
			if enforceValidation {
				log.Printf("Invalid client ID for user(id=%s), path=%s", domain.HashUserId(userId), c.Request.URL.Path)
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Invalid client ID",
					"message": "Client ID is invalid or expired. Call POST /api/v1/session/client-id to generate a new one.",
				})
				c.Abort()
				return
			} else {
				// Migration mode: just log warning
				log.Printf("WARNING: Invalid client ID for user(id=%s), path=%s (migration mode)", domain.HashUserId(userId), c.Request.URL.Path)
				c.Next()
				return
			}
		}

		// Valid client ID
		c.Next()
	}
}
