package deployment

import (
<<<<<<< HEAD
	"bytes"
	"encoding/json"
	"io"
=======
>>>>>>> main
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
)

func ComprehensiveLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Read request body (if any)
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// Restore the body so it can be read again
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		log.Println("╔════════════════════════════════════════════════════════════════")
		log.Printf("║ 📥 INCOMING REQUEST - %s", startTime.Format("2006-01-02 15:04:05"))
		log.Println("╠════════════════════════════════════════════════════════════════")

		// Basic Request Info
		log.Println("║ ⚙️  REQUEST DETAILS:")
		log.Printf("║   Method:        %s", c.Request.Method)
		log.Printf("║   URL:           %s", c.Request.URL.String())
		log.Printf("║   Path:          %s", c.Request.URL.Path)
		log.Printf("║   Query:         %s", c.Request.URL.RawQuery)
		log.Printf("║   Scheme:        %s", c.Request.URL.Scheme)
		log.Printf("║   Protocol:      %s", c.Request.Proto)
		log.Printf("║   Host:          %s", c.Request.Host)
		log.Printf("║   RemoteAddr:    %s", c.Request.RemoteAddr)
		log.Printf("║   ContentLength: %d", c.Request.ContentLength)
		log.Printf("║   TLS:           %v", c.Request.TLS != nil)

		// All Headers
		log.Println("║")
		log.Println("║ 📋 ALL HEADERS:")
		for name, values := range c.Request.Header {
			for _, value := range values {
				// Mask sensitive headers
				displayValue := value
				if strings.ToLower(name) == "authorization" ||
					strings.ToLower(name) == "cookie" {
					if len(value) > 20 {
						displayValue = value[:20] + "... [MASKED]"
					} else {
						displayValue = "[MASKED]"
					}
				}
				log.Printf("║   %s: %s", name, displayValue)
			}
		}

		// Important Proxy/HTTPS Headers
		log.Println("║")
		log.Println("║ 🔒 PROXY & HTTPS HEADERS:")
		log.Printf("║   X-Forwarded-Proto:  %s", c.GetHeader("X-Forwarded-Proto"))
		log.Printf("║   X-Forwarded-For:    %s", c.GetHeader("X-Forwarded-For"))
		log.Printf("║   X-Forwarded-Host:   %s", c.GetHeader("X-Forwarded-Host"))
		log.Printf("║   X-Real-IP:          %s", c.GetHeader("X-Real-IP"))
		log.Printf("║   X-Forwarded-Port:   %s", c.GetHeader("X-Forwarded-Port"))
		log.Printf("║   X-Forwarded-Scheme: %s", c.GetHeader("X-Forwarded-Scheme"))

		// CORS Headers
		log.Println("║")
		log.Println("║ 🌐 CORS HEADERS:")
		log.Printf("║   Origin:                        %s", c.GetHeader("Origin"))
		log.Printf("║   Referer:                       %s", c.GetHeader("Referer"))
		log.Printf("║   Access-Control-Request-Method: %s", c.GetHeader("Access-Control-Request-Method"))
		log.Printf("║   Access-Control-Request-Headers: %s", c.GetHeader("Access-Control-Request-Headers"))

		// Cookies
		log.Println("║")
		log.Println("║ 🍪 COOKIES:")
		cookies := c.Request.Cookies()
		if len(cookies) > 0 {
			for _, cookie := range cookies {
				log.Printf("║   %s = [MASKED] (Domain: %s, Path: %s, Secure: %v, HttpOnly: %v, SameSite: %v)",
					cookie.Name, cookie.Domain, cookie.Path, cookie.Secure, cookie.HttpOnly, cookie.SameSite)
			}
		} else {
			log.Println("║   (no cookies)")
		}

		// Request Body (if present and not too large)
		if len(bodyBytes) > 0 && len(bodyBytes) < 1024 {
			log.Println("║")
			log.Println("║ 📄 REQUEST BODY:")
			// Try to pretty print JSON
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, bodyBytes, "║   ", "  "); err == nil {
				log.Printf("║   %s", prettyJSON.String())
			} else {
				log.Printf("║   %s", string(bodyBytes))
			}
		} else if len(bodyBytes) > 0 {
			log.Printf("║ 📄 REQUEST BODY: %d bytes (too large to display)", len(bodyBytes))
		}

		log.Println("╚════════════════════════════════════════════════════════════════")

		// Process request
		c.Next()

		// Log response
		duration := time.Since(startTime)

		log.Println("╔════════════════════════════════════════════════════════════════")
		log.Printf("║ 📤 RESPONSE - Duration: %v", duration)
		log.Println("╠════════════════════════════════════════════════════════════════")
		log.Printf("║   Status Code: %d", c.Writer.Status())
		log.Printf("║   Size:        %d bytes", c.Writer.Size())

		// Response Headers
		log.Println("║")
		log.Println("║ 📋 RESPONSE HEADERS:")
		for name, values := range c.Writer.Header() {
			for _, value := range values {
				log.Printf("║   %s: %s", name, value)
			}
		}

		// Check if CORS headers are present
		log.Println("║")
		log.Println("║ 🌐 CORS RESPONSE HEADERS:")
		log.Printf("║   Access-Control-Allow-Origin:      %s", c.Writer.Header().Get("Access-Control-Allow-Origin"))
		log.Printf("║   Access-Control-Allow-Credentials: %s", c.Writer.Header().Get("Access-Control-Allow-Credentials"))
		log.Printf("║   Access-Control-Allow-Methods:     %s", c.Writer.Header().Get("Access-Control-Allow-Methods"))
		log.Printf("║   Access-Control-Allow-Headers:     %s", c.Writer.Header().Get("Access-Control-Allow-Headers"))

		log.Println("╚════════════════════════════════════════════════════════════════")
		log.Println()
	}
}

func GetGinRouter(corsConfiguration CorsConfiguration) *gin.Engine {
	_, err := ginmiddleware.OapiValidatorFromYamlFile("./openapi/api_v1.yaml")
	if err != nil {
		panic(err)
	}
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(ComprehensiveLogger())

	// ⭐ CORS configuration for subdomain cookie support
	router.Use(cors.New(
		cors.Config{
			AllowOriginFunc: func(origin string) bool {
				log.Default().Println("CORS request from:", origin)
				log.Default().Println("Current mode:", gin.Mode())
				log.Default().Println("Configured hostname:", corsConfiguration.Hostname)

				// Allow configured hostname
				if corsConfiguration.Hostname == "*" || corsConfiguration.Hostname == origin {
					return true
				}
				// Allow development origins only in non-production mode
				if gin.Mode() != gin.ReleaseMode {
					devOrigins := []string{
						"http://localhost:3000",
						"http://localhost:9002",
						"http://app.dailychexly.local.com:9002",
					}
					for _, devOrigin := range devOrigins {
						if origin == devOrigin {
							return true
						}
					}
				}
				return false
			},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Client-Id", "Cookie"},
			AllowCredentials: true, // ⭐ Enable cookies
			ExposeHeaders:    []string{"Content-Length", "Content-Type"},
			MaxAge:           12 * time.Hour,
		}))

	// router.Use(validator)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}
