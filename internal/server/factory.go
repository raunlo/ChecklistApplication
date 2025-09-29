package server

import (
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
}

func NewRoutes(
	engine *gin.Engine,
	checklistController checklistV1.IChecklistController,
	checklistItemController checklistItemV1.IChecklistItemController,
	sseController sse.ISSEController,
) IRoutes {
	return &routes{
		engine:                  engine,
		checklistController:     checklistController,
		checklistItemController: checklistItemController,
		sseController:           sseController,
	}
}

func (server *routes) ConfigureRoutes() {
	v1.RegisterV1Endpoints(server.engine.Group("/"),
		server.checklistController,
		server.checklistItemController,
		server.sseController)
}
