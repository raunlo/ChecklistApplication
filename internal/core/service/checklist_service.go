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
}

type checklistService struct {
	repository                repository.IChecklistRepository
	checklistOwnershipChecker guardrail.IChecklistOwnershipChecker
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
	return service.repository.DeleteChecklistById(ctx, id)
}

func (service *checklistService) FindAllChecklists(ctx context.Context) ([]domain.Checklist, domain.Error) {
	return service.repository.FindAllChecklists(ctx)
}
