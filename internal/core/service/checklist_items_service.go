package service

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	guardrail "com.raunlo.checklist/internal/core/guard_rail"
	"com.raunlo.checklist/internal/core/notification"
	"com.raunlo.checklist/internal/core/repository"
)

const (
	// MaxItemNameLength is the maximum allowed length for a checklist item name
	MaxItemNameLength = 500
	// MaxRowsPerItem is the maximum number of rows allowed per checklist item
	MaxRowsPerItem = 50
)

type IChecklistItemsService interface {
	SaveChecklistItem(context context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	UpdateChecklistItem(context context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	SaveChecklistItemRow(context context.Context, checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error)
	FindChecklistItemById(context context.Context, checklistId uint, id uint) (*domain.ChecklistItem, domain.Error)
	DeleteChecklistItemById(context context.Context, checklistId uint, id uint) domain.Error
	RestoreChecklistItem(context context.Context, checklistId uint, id uint) (domain.ChecklistItem, domain.Error)
	DeleteChecklistItemRow(context context.Context, checklistId uint, itemId uint, rowId uint) domain.Error
	FindAllChecklistItems(context context.Context, checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error)
	ChangeChecklistItemOrder(context context.Context, request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error)
	ToggleCompleted(context context.Context, checklistId uint, itemId uint, completed bool) (domain.ChecklistItem, domain.Error)
}

type checklistItemsService struct {
	repository                repository.IChecklistItemsRepository
	notifier                  notification.INotificationService
	checklistOwnershipChecker guardrail.IChecklistOwnershipChecker
	rebalanceService          IRebalanceService
}

func (service *checklistItemsService) UpdateChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return domain.ChecklistItem{}, err
	}

	// Validate item name length
	if len(checklistItem.Name) > MaxItemNameLength {
		return domain.ChecklistItem{}, domain.NewError("Item name exceeds maximum length of 500 characters", 400)
	}

	// Validate number of rows
	if len(checklistItem.Rows) > MaxRowsPerItem {
		return domain.ChecklistItem{}, domain.NewError("Item exceeds maximum of 50 rows", 400)
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

	// Validate item name length
	if len(checklistItem.Name) > MaxItemNameLength {
		return domain.ChecklistItem{}, domain.NewError("Item name exceeds maximum length of 500 characters", 400)
	}

	// Validate number of rows
	if len(checklistItem.Rows) > MaxRowsPerItem {
		return domain.ChecklistItem{}, domain.NewError("Item exceeds maximum of 50 rows", 400)
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

	// Validate that adding this row won't exceed the max rows limit
	item, err := service.repository.FindChecklistItemById(ctx, checklistId, itemId)
	if err != nil {
		return domain.ChecklistItemRow{}, err
	}
	if item == nil {
		return domain.ChecklistItemRow{}, domain.NewError("Checklist item not found", 404)
	}
	if len(item.Rows) >= MaxRowsPerItem {
		return domain.ChecklistItemRow{}, domain.NewError("Item has reached maximum of 50 rows", 400)
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
		// Notify with soft delete event (item can be restored)
		service.notifier.NotifyItemSoftDeleted(ctx, checklistId, id)
	}
	return err
}

func (service *checklistItemsService) RestoreChecklistItem(ctx context.Context, checklistId uint, id uint) (domain.ChecklistItem, domain.Error) {
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return domain.ChecklistItem{}, err
	}

	result, err := service.repository.RestoreChecklistItem(ctx, checklistId, id)
	if err == nil {
		service.notifier.NotifyItemRestored(ctx, checklistId, result)
	}
	return result, err
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
		// Trigger async rebalancing if gaps became too small
		if result.RebalanceNeeded && service.rebalanceService != nil {
			service.rebalanceService.TriggerRebalance(request.ChecklistId)
		}
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
