package deployment

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
)

type ctxKey string

const ctxKeyClientID ctxKey = "clientId"

func GetGinRouter(corsConfiguration CorsConfiguration) *gin.Engine {
	_, err := ginmiddleware.OapiValidatorFromYamlFile("./openapi/api_v1.yaml")
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

	// router.Use(validator)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}
