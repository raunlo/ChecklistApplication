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
	IsChecklistOwner(ctx context.Context, checklistId uint) domain.Error
}

type checklistOwnershipCheckerService struct {
	repository repository.IChecklistRepository
}

func (service *checklistOwnershipCheckerService) HasAccessToChecklist(ctx context.Context, checklistId uint) domain.Error {
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return err
	}

	hasAccess, err := service.repository.CheckUserHasAccessToChecklist(ctx, checklistId, userId)
	log.Printf("GuardRail: User(id=%s) access to checklist %d: %v", domain.GetHashedUserIdFromContext(ctx), checklistId, hasAccess)
	if err != nil {
		return domain.Wrap(err, "Failed to check user access to checklist", 500)
	}
	if !hasAccess {
		return error.NewChecklistNotFoundError(checklistId)
	}

	return nil
}

func (service *checklistOwnershipCheckerService) IsChecklistOwner(ctx context.Context, checklistId uint) domain.Error {
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return err
	}

	isOwner, err := service.repository.CheckUserIsOwner(ctx, checklistId, userId)
	log.Printf("GuardRail: User(id=%s) owner check for checklist %d: %v", domain.GetHashedUserIdFromContext(ctx), checklistId, isOwner)
	if err != nil {
		return domain.Wrap(err, "Failed to check checklist ownership", 500)
	}
	if !isOwner {
		return error.NewNotChecklistOwnerError(checklistId)
	}

	return nil
}

func NewChecklistOwnershipCheckerService(repository repository.IChecklistRepository) IChecklistOwnershipChecker {
	return &checklistOwnershipCheckerService{
		repository: repository,
	}
}
