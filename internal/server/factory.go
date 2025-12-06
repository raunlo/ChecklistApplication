package server

import (
	"time"

	"com.raunlo.checklist/internal/server/auth"
	v1 "com.raunlo.checklist/internal/server/v1"
	checklistV1 "com.raunlo.checklist/internal/server/v1/checklist"
	checklistItemV1 "com.raunlo.checklist/internal/server/v1/checklistItem"
	"com.raunlo.checklist/internal/server/v1/sse"
	"github.com/gin-gonic/gin"
)

type IRoutes interface {
	ConfigureRoutes()
}
type routes struct {
	engine                  *gin.Engine
	checklistController     checklistV1.IChecklistController
	checklistItemController checklistItemV1.IChecklistItemController
	sseController           sse.ISSEController
	googleSsoValidator      auth.IdtokenValidator
}

func NewRoutes(
	engine *gin.Engine,
	checklistController checklistV1.IChecklistController,
	checklistItemController checklistItemV1.IChecklistItemController,
	sseController sse.ISSEController,
	googleSsoValidator auth.IdtokenValidator,
) IRoutes {
	return &routes{
		engine:                  engine,
		checklistController:     checklistController,
		checklistItemController: checklistItemController,
		sseController:           sseController,
		googleSsoValidator:      googleSsoValidator,
	}
}

func (server *routes) ConfigureRoutes() {
	protectedGroup := server.engine.Group("/")
	protectedGroup.Use(auth.RateLimitMiddleware(100, time.Minute)) // 100 requests per minute
	protectedGroup.Use(auth.GoogleAuthMiddleware(server.googleSsoValidator))

	// Register all V1 endpoints with authentication middleware
	v1.RegisterV1Endpoints(protectedGroup,
		server.checklistController,
		server.checklistItemController,
		server.sseController)
}
