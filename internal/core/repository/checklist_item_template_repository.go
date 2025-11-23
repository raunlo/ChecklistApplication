package repository

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
)

type IChecklistItemTemplateRepository interface {
	SaveChecklistTemplate(ctx context.Context, checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error)
	GetAllChecklistTemplates(ctx context.Context) ([]domain.ChecklistItemTemplate, domain.Error)
	UpdateChecklistTemplate(ctx context.Context, checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error)
	DeleteChecklistTemplateById(ctx context.Context, id uint) domain.Error
	FindChecklistTemplateById(ctx context.Context, id uint) (*domain.ChecklistItemTemplate, domain.Error)
}
