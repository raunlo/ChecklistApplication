package repository

import "com.raunlo.checklist/internal/core/domain"

type IChecklistItemTemplateRepository interface {
	SaveChecklistTemplate(checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error)
	GetAllChecklistTemplates() ([]domain.ChecklistItemTemplate, domain.Error)
	UpdateChecklistTemplate(checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error)
	DeleteChecklistTemplateById(id uint) domain.Error
	FindChecklistTemplateById(id uint) (*domain.ChecklistItemTemplate, domain.Error)
}
