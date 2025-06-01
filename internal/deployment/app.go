package deployment

import (
	"fmt"

	"com.raunlo.checklist/internal/server"
	"github.com/gin-gonic/gin"
)

type Application struct {
	routes server.IRoutes
	router *gin.Engine
	config ServerConfiguration
}

func CreateApplication(routes server.IRoutes, router *gin.Engine, configuration ServerConfiguration) Application {
	return Application{
		routes: routes,
		router: router,
		config: configuration,
	}
}

func (application Application) StartApplication() error {
	application.routes.ConfigureRoutes()

	err := application.router.Run(fmt.Sprintf("localhost:%s", application.config.Port))
	return err
}
