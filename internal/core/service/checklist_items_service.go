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
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return err
	}
	err := service.repository.DeleteChecklistItemRow(ctx, checklistId, itemId, rowId)
	if err == nil {
		service.notifier.NotifyItemRowDeleted(ctx, checklistId, itemId, rowId)

		// After deleting a row, check if all remaining rows are completed
		// If so, auto-complete the parent item
		// This runs asynchronously to not block the delete response
		go func(bgChecklistId, bgItemId uint) {
			// Create a new background context that won't be cancelled when the request ends
			// Use context.Background() instead of the request context to avoid context cancellation issues
			bgCtx := context.Background()

			item, findErr := service.repository.FindChecklistItemById(bgCtx, bgChecklistId, bgItemId)
			if findErr != nil || item == nil || item.Completed {
				return // Nothing to do
			}

			// Early exit: if any row is not completed, no need to update
			if len(item.Rows) == 0 {
				return // No rows left, nothing to update
			}

			for _, row := range item.Rows {
				if !row.Completed {
					return // Found incomplete row, exit early
				}
			}

			// All remaining rows are completed, mark parent as completed
			item.Completed = true
			_, updateErr := service.repository.UpdateChecklistItem(bgCtx, bgChecklistId, *item)
			if updateErr == nil {
				// Notify about the automatic completion
				service.notifier.NotifyItemUpdated(bgCtx, bgChecklistId, *item)
			}
			// Errors are silently ignored since the delete operation already succeeded
		}(checklistId, itemId)
	}
	return err
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
