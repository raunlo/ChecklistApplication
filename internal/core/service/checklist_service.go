package service

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/error"
	guardrail "com.raunlo.checklist/internal/core/guard_rail"
	"com.raunlo.checklist/internal/core/repository"
)

type IChecklistService interface {
	UpdateChecklist(ctx context.Context, checklist domain.Checklist) (domain.Checklist, domain.Error)
	SaveChecklist(ctx context.Context, checklist domain.Checklist) (domain.Checklist, domain.Error)
	FindChecklistById(ctx context.Context, id uint) (*domain.Checklist, domain.Error)
	DeleteChecklistById(ctx context.Context, id uint) domain.Error
	FindAllChecklists(ctx context.Context) ([]domain.Checklist, domain.Error)
	LeaveSharedChecklist(ctx context.Context, checklistId uint) domain.Error
}

type checklistService struct {
	repository                repository.IChecklistRepository
	checklistOwnershipChecker guardrail.IChecklistOwnershipChecker
	checklistItemService      IChecklistItemsService
}

func (service *checklistService) UpdateChecklist(ctx context.Context, checklist domain.Checklist) (domain.Checklist, domain.Error) {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklist.Id); err != nil {
		return domain.Checklist{}, error.NewChecklistNotFoundError(checklist.Id)
	}
	return service.repository.UpdateChecklist(ctx, checklist)
}

func (service *checklistService) SaveChecklist(ctx context.Context, checklist domain.Checklist) (domain.Checklist, domain.Error) {
	return service.repository.SaveChecklist(ctx, checklist)
}

func (service *checklistService) FindChecklistById(ctx context.Context, id uint) (*domain.Checklist, domain.Error) {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, id); err != nil {
		return nil, error.NewChecklistNotFoundError(id)
	}
	return service.repository.FindChecklistById(ctx, id)
}

func (service *checklistService) DeleteChecklistById(ctx context.Context, id uint) domain.Error {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, id); err != nil {
		return error.NewChecklistNotFoundError(id)
	}

	if checklistItems, err := service.checklistItemService.FindAllChecklistItems(ctx, id, nil, domain.AscSort); err != nil {
		return err
	} else if len(checklistItems) > 0 {
		// If there are items in the checklist, we can't delete it
		return domain.NewError("Checklist is not empty", 400)
	}
	return service.repository.DeleteChecklistById(ctx, id)
}

func (service *checklistService) FindAllChecklists(ctx context.Context) ([]domain.Checklist, domain.Error) {
	return service.repository.FindAllChecklists(ctx)
}

func (service *checklistService) LeaveSharedChecklist(ctx context.Context, checklistId uint) domain.Error {
	// Get userId from context
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return err
	}

	// Guard rail: Check if user has access to this checklist
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return error.NewChecklistNotFoundError(checklistId)
	}

	// Guard rail: Prevent owner from leaving their own checklist
	if err := service.checklistOwnershipChecker.IsChecklistOwner(ctx, checklistId); err == nil {
		return domain.NewError("Checklist owners cannot leave their own checklists. Delete the checklist instead.", 400)
	}

	// Delete the share (remove user's access)
	return service.repository.DeleteChecklistShare(ctx, checklistId, userId)
}
