package deployment

import (
	"net/http"
	"strings"
	"time"

	"com.raunlo.checklist/internal/server/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
)

func GetGinRouter(corsConfiguration CorsConfiguration) *gin.Engine {
	_, err := ginmiddleware.OapiValidatorFromYamlFile("./openapi/api_v1.yaml")
	if err != nil {
		panic(err)
	}
	router := gin.New()
	router.Use(middleware.SecurityHeadersMiddleware()) // Security headers first
	router.Use(gin.Logger())

	// Parse comma-separated configured origins
	allowedOrigins := make(map[string]bool)
	for _, origin := range strings.Split(corsConfiguration.Hostname, ",") {
		trimmedOrigin := strings.TrimSpace(origin)
		if trimmedOrigin != "" {
			allowedOrigins[trimmedOrigin] = true
		}
	}

	// Add development origins only in non-release mode
	if gin.Mode() != gin.ReleaseMode {
		devOrigins := []string{
			"http://localhost:3000",
			"http://localhost:9002",
			"http://app.dailychexly.local.com:9002",
		}
		for _, devOrigin := range devOrigins {
			allowedOrigins[devOrigin] = true
		}
	}

	// ⭐ CORS configuration with strict origin validation
	router.Use(cors.New(
		cors.Config{
			AllowOriginFunc: func(origin string) bool {
				// Check if origin is in allowed list
				return allowedOrigins[origin]
			},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Client-Id", "X-CSRF-Token", "Cookie"},
			AllowCredentials: true, // ⭐ Enable cookies
			ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-CSRF-Token"},
			MaxAge:           12 * time.Hour,
		}))

	// router.Use(validator)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}
