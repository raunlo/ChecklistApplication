package service

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	guardrail "com.raunlo.checklist/internal/core/guard_rail"
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
	repository                repository.IChecklistItemsRepository
	notifier                  notification.INotificationService
	checklistOwnershipChecker guardrail.IChecklistOwnershipChecker
}

func (service *checklistItemsService) UpdateChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return domain.ChecklistItem{}, err
	}

	result, err := service.repository.UpdateChecklistItem(ctx, checklistId, checklistItem)
	if err == nil {
		service.notifier.NotifyItemUpdated(ctx, checklistId, result)
	}
	return result, err
}

func (service *checklistItemsService) SaveChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return domain.ChecklistItem{}, err
	}
	result, err := service.repository.SaveChecklistItem(ctx, checklistId, checklistItem)
	if err == nil {
		service.notifier.NotifyItemCreated(ctx, checklistId, result)
	}
	return result, err
}

func (service *checklistItemsService) SaveChecklistItemRow(ctx context.Context, checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error) {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return domain.ChecklistItemRow{}, err
	}
	result, err := service.repository.SaveChecklistItemRow(ctx, checklistId, itemId, row)
	if err == nil {
		service.notifier.NotifyItemRowAdded(ctx, checklistId, itemId, result)
	}
	return result, err
}

func (service *checklistItemsService) FindChecklistItemById(ctx context.Context, checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return nil, err
	}
	return service.repository.FindChecklistItemById(ctx, checklistId, id)
}

func (service *checklistItemsService) DeleteChecklistItemById(ctx context.Context, checklistId uint, id uint) domain.Error {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return err
	}

	err := service.repository.DeleteChecklistItemById(ctx, checklistId, id)
	if err == nil {
		service.notifier.NotifyItemDeleted(ctx, checklistId, id)
	}
	return err
}

func (service *checklistItemsService) DeleteChecklistItemRow(ctx context.Context, checklistId uint, itemId uint, rowId uint) domain.Error {
	// Auth check: Verify user has access to this checklist before any operations
	// This ensures the subsequent transaction operations are authorized
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return err
	}

	// The repository handles row deletion and auto-completion atomically in a single transaction
	// This prevents race conditions when multiple rows are deleted concurrently
	// The SQL ensures checklistId is validated in all queries, preventing unauthorized access
	_, err := service.repository.DeleteChecklistItemRowAndAutoComplete(ctx, checklistId, itemId, rowId)
	if err != nil {
		return err
	}

	// Notify about row deletion
	service.notifier.NotifyItemRowDeleted(ctx, checklistId, itemId, rowId)

	return nil
}

func (service *checklistItemsService) FindAllChecklistItems(ctx context.Context, checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error) {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return nil, err
	}

	return service.repository.FindAllChecklistItems(ctx, checklistId, completed, sortOrder)
}

func (service *checklistItemsService) ChangeChecklistItemOrder(ctx context.Context, request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error) {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, request.ChecklistId); err != nil {
		return domain.ChangeOrderResponse{}, err
	}
	result, err := service.repository.ChangeChecklistItemOrder(ctx, request)
	if err == nil {
		service.notifier.NotifyItemReordered(ctx, request, result)
	}
	return result, err
}

func (service *checklistItemsService) ToggleCompleted(ctx context.Context, checklistId uint, itemId uint, completed bool) (domain.ChecklistItem, domain.Error) {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return domain.ChecklistItem{}, err
	}
	result, err := service.repository.ToggleItemCompleted(ctx, checklistId, itemId, completed)
	if err == nil {
		service.notifier.NotifyItemUpdated(ctx, checklistId, result)
	}
	return result, err
}
