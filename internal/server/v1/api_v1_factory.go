package v1

import (
	"com.raunlo.checklist/internal/server/v1/checklist"
	"com.raunlo.checklist/internal/server/v1/checklistItem"
	"com.raunlo.checklist/internal/server/v1/sse"
	"github.com/gin-gonic/gin"
)

func RegisterV1Endpoints(gin *gin.RouterGroup,
	checklistController checklist.IChecklistController,
	checklistItemController checklistItem.IChecklistItemController,
	sseController sse.ISSEController,
) {
	checklist.RegisterHandlers(gin, checklist.NewStrictHandler(checklistController, nil))
	checklistItem.RegisterHandlers(gin, checklistItem.NewStrictHandler(checklistItemController, nil))
	sse.RegisterHandlers(gin, sse.NewStrictHandler(sseController, nil))
}
