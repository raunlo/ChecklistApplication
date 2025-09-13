package deployment

import (
	"net/http"
	"time"

	"com.raunlo.checklist/internal/sse"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
)

func GetGinRouter(corsConfiguration CorsConfiguration) *gin.Engine {

	validator, err := ginmiddleware.OapiValidatorFromYamlFile("./openapi/api_v1.yaml")
	if err != nil {
		panic(err)
	}
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(cors.New(
		cors.Config{
			AllowOrigins:     []string{corsConfiguration.Hostname},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"*"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))

	router.Use(func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent) // 204
			c.Abort()
			return
		}
		c.Next()
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Apply OpenAPI validator to all routes except SSE endpoint
	router.Use(func(c *gin.Context) {
		// Skip validation for SSE endpoint
		if c.Request.URL.Path == "/events" {
			c.Next()
			return
		}
		// Apply validator for all other routes
		validator(c)
	})

	// SSE endpoint for clients to subscribe to events
	router.GET("/events", gin.WrapH(sse.DefaultBroker.Handler()))

	return router
}
