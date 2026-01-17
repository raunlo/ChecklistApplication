package middleware

import "github.com/gin-gonic/gin"

// SecurityHeadersMiddleware adds security headers to all responses
// Protects against common web vulnerabilities (clickjacking, XSS, MIME sniffing, etc.)
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking attacks
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME-type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection (legacy but still good to have)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (strict - only allow resources from same origin)
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self'; "+
				"connect-src 'self'; "+
				"img-src 'self' data:; "+
				"style-src 'self' 'unsafe-inline'; "+
				"base-uri 'self'; "+
				"form-action 'self'")

		// HTTP Strict Transport Security (only in production)
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Permissions Policy (disable unnecessary browser features)
		c.Header("Permissions-Policy",
			"geolocation=(), "+
				"microphone=(), "+
				"camera=(), "+
				"payment=(), "+
				"usb=(), "+
				"magnetometer=(), "+
				"gyroscope=()")

		c.Next()
	}
}
