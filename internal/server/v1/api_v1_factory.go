package v1

import (
	"com.raunlo.checklist/internal/server/v1/checklist"
	"com.raunlo.checklist/internal/server/v1/checklistItem"
	"com.raunlo.checklist/internal/server/v1/session"
	"com.raunlo.checklist/internal/server/v1/sse"
	"com.raunlo.checklist/internal/server/v1/user"
	"github.com/gin-gonic/gin"
)

type V1ProtectedEndpointsRegistrationRequest struct {
	ChecklistController     checklist.IChecklistController
	ChecklistItemController checklistItem.IChecklistItemController
	SSEController           sse.ISSEController
	UserController          user.IUserController
	SessionController       session.ISessionController
}

func RegisterV1Endpoints(gin *gin.RouterGroup,
	request V1ProtectedEndpointsRegistrationRequest,
) {
	checklist.RegisterHandlers(gin, checklist.NewStrictHandler(request.ChecklistController, nil))
	checklistItem.RegisterHandlers(gin, checklistItem.NewStrictHandler(request.ChecklistItemController, nil))
	sse.RegisterHandlers(gin, sse.NewStrictHandler(request.SSEController, nil))
	user.RegisterHandlers(gin, user.NewStrictHandler(request.UserController, nil))
	session.RegisterHandlers(gin, session.NewStrictHandler(request.SessionController, nil))
}
