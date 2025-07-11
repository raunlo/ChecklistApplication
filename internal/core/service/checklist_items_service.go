package service

import (
	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/repository"
)

type IChecklistItemsService interface {
	SaveChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	UpdateChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	FindChecklistItemById(checklistId uint, id uint) (*domain.ChecklistItem, domain.Error)
	DeleteChecklistItemById(checklistId uint, id uint) domain.Error
	FindAllChecklistItems(checklistId uint, completed *bool, sortOrder *domain.SortOrder) ([]domain.ChecklistItem, domain.Error)
	ChangeChecklistItemOrder(request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error)
}

type checklistItemsService struct {
	repository      repository.IChecklistItemsRepository
	templateService IChecklistItemTemplateService
}

func (service *checklistItemsService) UpdateChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	checklistItem.Completed = isChecklistItemCompleted(checklistItem)
	return service.repository.UpdateChecklistItem(checklistId, checklistItem)
}

func (service *checklistItemsService) SaveChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	//checklistItem.Completed = isChecklistItemCompleted(checklistItem)
	return service.repository.SaveChecklistItem(checklistId, checklistItem)
}

func (service *checklistItemsService) FindChecklistItemById(checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
	return service.repository.FindChecklistItemById(checklistId, id)
}

func (service *checklistItemsService) DeleteChecklistItemById(checklistId uint, id uint) domain.Error {
	return service.repository.DeleteChecklistItemById(checklistId, id)
}

func (service *checklistItemsService) FindAllChecklistItems(checklistId uint, completed *bool, sortOrder *domain.SortOrder) ([]domain.ChecklistItem, domain.Error) {
	return service.repository.FindAllChecklistItems(checklistId, false, domain.AscSort)
}

func (service *checklistItemsService) ChangeChecklistItemOrder(request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error) {
	return service.repository.ChangeChecklistItemOrder(request)
}

func isChecklistItemCompleted(checklistItem domain.ChecklistItem) bool {
	for _, row := range checklistItem.Rows {
		if !row.Completed {
			return false
		}
	}
	return true
}
