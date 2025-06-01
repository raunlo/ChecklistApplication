package deployment

import (
	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
)

func GetGinRouter() *gin.Engine {

	validator, err := ginmiddleware.OapiValidatorFromYamlFile("./openapi/api_v1.yaml")
	if err != nil {
		panic(err)
	}
	router := gin.New()
	router.Use(validator)
	router.Use(gin.Logger())
	return router
}
