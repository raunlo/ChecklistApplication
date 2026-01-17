package session

import (
	"context"
	"log"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/service"
)

type ISessionController interface {
	StrictServerInterface
}

type sessionControllerImpl struct {
	sessionService service.ISessionService
}

func NewSessionController(sessionService service.ISessionService) ISessionController {
	return &sessionControllerImpl{
		sessionService: sessionService,
	}
}

// GenerateClientId generates a server-validated client ID for SSE connections
func (ctrl *sessionControllerImpl) GenerateClientId(ctx context.Context, request GenerateClientIdRequestObject) (GenerateClientIdResponseObject, error) {
	userId, _ := domain.GetUserIdFromContext(ctx)

	clientId, err := ctrl.sessionService.GenerateClientId(userId)
	if err != nil {
		log.Printf("Failed to generate client ID for user(id=%s): %v", domain.GetHashedUserIdFromContext(ctx), err)
		return GenerateClientId500JSONResponse{
			Message: "Failed to generate client ID",
		}, nil
	}

	log.Printf("Generated client ID for user(id=%s)", domain.GetHashedUserIdFromContext(ctx))

	return GenerateClientId200JSONResponse{
		ClientId: clientId,
	}, nil
}
