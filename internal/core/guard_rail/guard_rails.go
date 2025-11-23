package guardrail

import (
	"context"
	"log"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/error"
	"com.raunlo.checklist/internal/core/repository"
)

type IChecklistOwnershipChecker interface {
	HasAccessToChecklist(ctx context.Context, checklistId uint) domain.Error
}

type checklistOwnershipCheckerService struct {
	repository repository.IChecklistRepository
}

func (service *checklistOwnershipCheckerService) HasAccessToChecklist(ctx context.Context, checklistId uint) domain.Error {
	userId := ctx.Value(domain.UserIdContextKey).(string)
	hasAccess, err := service.repository.CheckUserHasAccessToChecklist(checklistId, userId)
	log.Printf("GuardRail: User(id=%s) access to checklist %d: %v", userId, checklistId, hasAccess)
	if err != nil {
		return domain.Wrap(err, "Failed to check user access to checklist", 500)
	}
	if !hasAccess {
		return error.NewChecklistNotFoundError(checklistId)
	}

	return nil
}

func NewChecklistOwnershipCheckerService(repository repository.IChecklistRepository) IChecklistOwnershipChecker {
	return &checklistOwnershipCheckerService{
		repository: repository,
	}
}
