package service

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/notification"
	"com.raunlo.checklist/internal/core/repository"
)

type IChecklistItemsService interface {
	SaveChecklistItem(context context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	UpdateChecklistItem(context context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	SaveChecklistItemRow(context context.Context, checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error)
	FindChecklistItemById(context context.Context, checklistId uint, id uint) (*domain.ChecklistItem, domain.Error)
	DeleteChecklistItemById(context context.Context, checklistId uint, id uint) domain.Error
	DeleteChecklistItemRow(context context.Context, checklistId uint, itemId uint, rowId uint) domain.Error
	FindAllChecklistItems(context context.Context, checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error)
	ChangeChecklistItemOrder(context context.Context, request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error)
	ToggleCompleted(context context.Context, checklistId uint, itemId uint, completed bool) (domain.ChecklistItem, domain.Error)
}

type checklistItemsService struct {
	repository      repository.IChecklistItemsRepository
	templateService IChecklistItemTemplateService
	notifier        notification.INotificationService
}

func (service *checklistItemsService) UpdateChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	//checklistItem.Completed = isChecklistItemCompleted(checklistItem)
	result, err := service.repository.UpdateChecklistItem(checklistId, checklistItem)
	if err == nil {
		service.notifier.NotifyItemUpdated(ctx, checklistId, result)
	}
	return result, err
}

func (service *checklistItemsService) SaveChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	//checklistItem.Completed = isChecklistItemCompleted(checklistItem)
	result, err := service.repository.SaveChecklistItem(checklistId, checklistItem)
	if err == nil {
		service.notifier.NotifyItemCreated(ctx, checklistId, result)
	}
	return result, err
}

func (service *checklistItemsService) SaveChecklistItemRow(ctx context.Context, checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error) {
	result, err := service.repository.SaveChecklistItemRow(checklistId, itemId, row)
	if err == nil {
		service.notifier.NotifyItemRowAdded(ctx, checklistId, itemId, result)
	}
	return result, err
}

func (service *checklistItemsService) FindChecklistItemById(ctx context.Context, checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
	return service.repository.FindChecklistItemById(checklistId, id)
}

func (service *checklistItemsService) DeleteChecklistItemById(ctx context.Context, checklistId uint, id uint) domain.Error {
	err := service.repository.DeleteChecklistItemById(checklistId, id)
	if err == nil {
		service.notifier.NotifyItemDeleted(ctx, checklistId, id)
	}
	return err
}

func (service *checklistItemsService) DeleteChecklistItemRow(ctx context.Context, checklistId uint, itemId uint, rowId uint) domain.Error {
	err := service.repository.DeleteChecklistItemRow(checklistId, itemId, rowId)
	if err == nil {
		service.notifier.NotifyItemRowDeleted(ctx, checklistId, itemId, rowId)
	}
	return err
}

func (service *checklistItemsService) FindAllChecklistItems(ctx context.Context, checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error) {
	return service.repository.FindAllChecklistItems(checklistId, completed, sortOrder)
}

func (service *checklistItemsService) ChangeChecklistItemOrder(ctx context.Context, request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error) {
	result, err := service.repository.ChangeChecklistItemOrder(request)
	if err == nil {
		service.notifier.NotifyItemReordered(ctx, request, result)
	}
	return result, err
}

func (service *checklistItemsService) ToggleCompleted(ctx context.Context, checklistId uint, itemId uint, completed bool) (domain.ChecklistItem, domain.Error) {
	result, err := service.repository.ToggleItemCompleted(checklistId, itemId, completed)
	if err == nil {
		service.notifier.NotifyItemUpdated(ctx, checklistId, result)
	}
	return result, err
}
