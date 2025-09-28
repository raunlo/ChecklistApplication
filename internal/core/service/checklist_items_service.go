package service

import (
	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/notification"
	"com.raunlo.checklist/internal/core/repository"
)

type IChecklistItemsService interface {
	SaveChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	UpdateChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	SaveChecklistItemRow(checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error)
	FindChecklistItemById(checklistId uint, id uint) (*domain.ChecklistItem, domain.Error)
	DeleteChecklistItemById(checklistId uint, id uint) domain.Error
	DeleteChecklistItemRow(checklistId uint, itemId uint, rowId uint) domain.Error
	FindAllChecklistItems(checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error)
	ChangeChecklistItemOrder(request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error)
	ToggleCompleted(checklistId uint, itemId uint, completed bool) (domain.ChecklistItem, domain.Error)
}

type checklistItemsService struct {
	repository      repository.IChecklistItemsRepository
	templateService IChecklistItemTemplateService
	notifier        notification.INotificationService
}

func (service *checklistItemsService) UpdateChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	//checklistItem.Completed = isChecklistItemCompleted(checklistItem)
	result, err := service.repository.UpdateChecklistItem(checklistId, checklistItem)
	if err == nil {
		service.notifier.NotifyItemUpdated(checklistId, result)
	}
	return result, err
}

func (service *checklistItemsService) SaveChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	//checklistItem.Completed = isChecklistItemCompleted(checklistItem)
	result, err := service.repository.SaveChecklistItem(checklistId, checklistItem)
	if err == nil {
		service.notifier.NotifyItemCreated(checklistId, result)
	}
	return result, err
}

func (service *checklistItemsService) SaveChecklistItemRow(checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error) {
	result, err := service.repository.SaveChecklistItemRow(checklistId, itemId, row)
	if err == nil {
		service.notifier.NotifyItemRowAdded(checklistId, itemId, result)
	}
	return result, err
}

func (service *checklistItemsService) FindChecklistItemById(checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
	return service.repository.FindChecklistItemById(checklistId, id)
}

func (service *checklistItemsService) DeleteChecklistItemById(checklistId uint, id uint) domain.Error {
	err := service.repository.DeleteChecklistItemById(checklistId, id)
	if err == nil {
		service.notifier.NotifyItemDeleted(checklistId, id)
	}
	return err
}

func (service *checklistItemsService) DeleteChecklistItemRow(checklistId uint, itemId uint, rowId uint) domain.Error {
	err := service.repository.DeleteChecklistItemRow(checklistId, itemId, rowId)
	if err == nil {
		service.notifier.NotifyItemRowDeleted(checklistId, itemId, rowId)
	}
	return err
}

func (service *checklistItemsService) FindAllChecklistItems(checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error) {
	return service.repository.FindAllChecklistItems(checklistId, completed, sortOrder)
}

func (service *checklistItemsService) ChangeChecklistItemOrder(request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error) {
	result, err := service.repository.ChangeChecklistItemOrder(request)
	if err == nil {
		service.notifier.NotifyItemReordered(request, result)
	}
	return result, err
}

func (service *checklistItemsService) ToggleCompleted(checklistId uint, itemId uint, completed bool) (domain.ChecklistItem, domain.Error) {
	result, err := service.repository.ToggleItemCompleted(checklistId, itemId, completed)
	if err == nil {
		service.notifier.NotifyItemUpdated(checklistId, result)
	}
	return result, err
}
