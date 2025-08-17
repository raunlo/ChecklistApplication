package v1

import (
	"com.raunlo.checklist/internal/server/v1/checklist"
	"com.raunlo.checklist/internal/server/v1/checklistItem"
	checklistItemTemplate "com.raunlo.checklist/internal/server/v1/checklistItemTemplate"
	"github.com/gin-gonic/gin"
)

func RegisterV1Endpoints(gin *gin.RouterGroup,
	checklistController checklist.IChecklistController,
	checklistItemController checklistItem.IChecklistItemController,
	checklistItemTemplateController checklistItemTemplate.IChecklistItemTemplateController) {
	checklist.RegisterHandlers(gin, checklist.NewStrictHandler(checklistController, nil))
	checklistItem.RegisterHandlers(gin, checklistItem.NewStrictHandler(checklistItemController, nil))
	checklistItemTemplate.RegisterHandlers(gin, checklistItemTemplate.NewStrictHandler(checklistItemTemplateController, nil))
}
