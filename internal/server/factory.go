package server

import (
	v1 "com.raunlo.checklist/internal/server/v1"
	checklistV1 "com.raunlo.checklist/internal/server/v1/checklist"
	checklistItemV1 "com.raunlo.checklist/internal/server/v1/checklistItem"
	checklistItemTemplateV1 "com.raunlo.checklist/internal/server/v1/checklistItemTemplate"
	"github.com/gin-gonic/gin"
)

type IRoutes interface {
	ConfigureRoutes()
}
type routes struct {
	engine                          *gin.Engine
	checklistController             checklistV1.IChecklistController
	checklistItemController         checklistItemV1.IChecklistItemController
	checklistItemTemplateController checklistItemTemplateV1.IChecklistItemTemplateController
}

func NewRoutes(
	engine *gin.Engine,
	checklistController checklistV1.IChecklistController,
	checklistItemController checklistItemV1.IChecklistItemController,
	checklistItemTemplateController checklistItemTemplateV1.IChecklistItemTemplateController) IRoutes {

	return &routes{
		engine:                          engine,
		checklistController:             checklistController,
		checklistItemController:         checklistItemController,
		checklistItemTemplateController: checklistItemTemplateController,
	}
}

func (server *routes) ConfigureRoutes() {
	v1.RegisterV1Endpoints(server.engine.Group("/"),
		server.checklistController,
		server.checklistItemController,
		server.checklistItemTemplateController)
}
