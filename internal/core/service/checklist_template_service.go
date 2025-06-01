package service

import (
	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/repository"
)

type IChecklistItemTemplateService interface {
	SaveChecklistTemplate(checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error)
	GetAllChecklistTemplates() ([]domain.ChecklistItemTemplate, domain.Error)
	UpdateChecklistTemplate(checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error)
	DeleteChecklistTemplateById(id uint) domain.Error
	FindChecklistTemplateById(id uint) (*domain.ChecklistItemTemplate, domain.Error)
}

type checklistItemTemplateService struct {
	repository repository.IChecklistItemTemplateRepository
}

func (service *checklistItemTemplateService) SaveChecklistTemplate(checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error) {
	return service.repository.SaveChecklistTemplate(checklistTemplate)
}

func (service *checklistItemTemplateService) GetAllChecklistTemplates() ([]domain.ChecklistItemTemplate, domain.Error) {
	return service.repository.GetAllChecklistTemplates()
}

func (service *checklistItemTemplateService) UpdateChecklistTemplate(checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error) {
	return service.repository.UpdateChecklistTemplate(checklistTemplate)
}

func (service *checklistItemTemplateService) DeleteChecklistTemplateById(id uint) domain.Error {
	return service.repository.DeleteChecklistTemplateById(id)
}

func (service *checklistItemTemplateService) FindChecklistTemplateById(id uint) (*domain.ChecklistItemTemplate, domain.Error) {
	return service.repository.FindChecklistTemplateById(id)
}
