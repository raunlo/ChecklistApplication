package deployment

import (
	"net/http"
	"time"

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
	router.Use(validator)
	router.Use(gin.Logger())
	router.Use(cors.New(
		cors.Config{
			AllowOrigins:  []string{corsConfiguration.Hostname},
			AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:  []string{"Origin", "Content-Type", "Authorization", "Accept"},
			ExposeHeaders: []string{"Content-Length"},
			//AllowCredentials: true,
			MaxAge: 12 * time.Hour,
		}))

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return router
}
