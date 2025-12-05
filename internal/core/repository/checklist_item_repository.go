package repository

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
)

type IChecklistItemsRepository interface {
	UpdateChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	SaveChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	SaveChecklistItemRow(ctx context.Context, checklistId uint, checklistItemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error)
	FindChecklistItemById(ctx context.Context, checklistId uint, id uint) (*domain.ChecklistItem, domain.Error)
	DeleteChecklistItemById(ctx context.Context, checklistId uint, id uint) domain.Error
	// DeleteChecklistItemRowAndAutoComplete atomically deletes a row and auto-completes the parent item if all remaining rows are completed
	DeleteChecklistItemRowAndAutoComplete(ctx context.Context, checklistId uint, itemId uint, rowId uint) domain.Error
	FindAllChecklistItems(ctx context.Context, checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error)
	ChangeChecklistItemOrder(ctx context.Context, request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error)
	ToggleItemCompleted(ctx context.Context, checklistId uint, checklistItemId uint, completed bool) (domain.ChecklistItem, domain.Error)
}
