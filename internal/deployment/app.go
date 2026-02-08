package deployment

import (
	"fmt"

	"com.raunlo.checklist/internal/job"
	"com.raunlo.checklist/internal/server"
	"github.com/gin-gonic/gin"
)

type Application struct {
	routes     server.IRoutes
	router     *gin.Engine
	config     ServerConfiguration
	cleanupJob *job.CleanupJob
}

func CreateApplication(routes server.IRoutes, router *gin.Engine, configuration ServerConfiguration, cleanupJob *job.CleanupJob) Application {
	return Application{
		routes:     routes,
		router:     router,
		config:     configuration,
		cleanupJob: cleanupJob,
	}
}

func (application Application) StartApplication() error {
	application.routes.ConfigureRoutes()

	// Start background jobs
	if application.cleanupJob != nil {
		application.cleanupJob.Start()
	}

	err := application.router.Run(fmt.Sprintf(":%s", application.config.Port))
	return err
}
