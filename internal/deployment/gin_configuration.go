package deployment

import (
	"net/http"
	"time"

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
	router.Use(gin.Logger())

	// ⭐ CORS configuration for subdomain cookie support
	router.Use(cors.New(
		cors.Config{
			AllowOriginFunc: func(origin string) bool {
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
