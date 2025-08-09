package repository

import "com.raunlo.checklist/internal/core/domain"

type IChecklistItemsRepository interface {
	UpdateChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	SaveChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	SaveChecklistItemRow(checklistId uint, checklistItemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error)
	FindChecklistItemById(checklistId uint, id uint) (*domain.ChecklistItem, domain.Error)
	DeleteChecklistItemById(checklistId uint, id uint) domain.Error
	DeleteChecklistItemRow(checklistId uint, itemId uint, rowId uint) domain.Error
	FindAllChecklistItems(checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error)
	ChangeChecklistItemOrder(request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error)
}
