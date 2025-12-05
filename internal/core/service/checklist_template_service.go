package service

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/repository"
)

type IChecklistItemTemplateService interface {
	SaveChecklistTemplate(ctx context.Context, checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error)
	GetAllChecklistTemplates(ctx context.Context) ([]domain.ChecklistItemTemplate, domain.Error)
	UpdateChecklistTemplate(ctx context.Context, checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error)
	DeleteChecklistTemplateById(ctx context.Context, id uint) domain.Error
	FindChecklistTemplateById(ctx context.Context, id uint) (*domain.ChecklistItemTemplate, domain.Error)
}

type checklistItemTemplateService struct {
	repository repository.IChecklistItemTemplateRepository
}

func (service *checklistItemTemplateService) SaveChecklistTemplate(ctx context.Context, checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error) {
	return service.repository.SaveChecklistTemplate(ctx, checklistTemplate)
}

func (service *checklistItemTemplateService) GetAllChecklistTemplates(ctx context.Context) ([]domain.ChecklistItemTemplate, domain.Error) {
	return service.repository.GetAllChecklistTemplates(ctx)
}

func (service *checklistItemTemplateService) UpdateChecklistTemplate(ctx context.Context, checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error) {
	return service.repository.UpdateChecklistTemplate(ctx, checklistTemplate)
}

func (service *checklistItemTemplateService) DeleteChecklistTemplateById(ctx context.Context, id uint) domain.Error {
	return service.repository.DeleteChecklistTemplateById(ctx, id)
}

func (service *checklistItemTemplateService) FindChecklistTemplateById(ctx context.Context, id uint) (*domain.ChecklistItemTemplate, domain.Error) {
	return service.repository.FindChecklistTemplateById(ctx, id)
}
