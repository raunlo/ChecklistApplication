package repository

import (
	"context"
	"time"

	"com.raunlo.checklist/internal/core/domain"
)

type IChecklistItemsRepository interface {
	UpdateChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	SaveChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	SaveChecklistItemRow(ctx context.Context, checklistId uint, checklistItemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error)
	FindChecklistItemById(ctx context.Context, checklistId uint, id uint) (*domain.ChecklistItem, domain.Error)
	DeleteChecklistItemById(ctx context.Context, checklistId uint, id uint) domain.Error
	// DeleteChecklistItemRowAndAutoComplete atomically deletes a row and auto-completes the parent item if all remaining rows are completed
	// Returns a result indicating whether the deletion was successful and if auto-completion occurred
	DeleteChecklistItemRowAndAutoComplete(ctx context.Context, checklistId uint, itemId uint, rowId uint) (domain.ChecklistItemRowDeletionResult, domain.Error)
	FindAllChecklistItems(ctx context.Context, checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error)
	ChangeChecklistItemOrder(ctx context.Context, request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error)
	ToggleItemCompleted(ctx context.Context, checklistId uint, checklistItemId uint, completed bool) (domain.ChecklistItem, domain.Error)
	// RebalancePositions redistributes positions evenly for all items in a checklist
	RebalancePositions(ctx context.Context, checklistId uint) domain.Error
	// RestoreChecklistItem restores a soft-deleted item (undo functionality)
	RestoreChecklistItem(ctx context.Context, checklistId uint, itemId uint) (domain.ChecklistItem, domain.Error)
	// PurgeSoftDeletedItems permanently deletes items that were soft-deleted before the retention period
	// Returns the number of items purged
	PurgeSoftDeletedItems(ctx context.Context, retentionPeriod time.Duration) (int64, domain.Error)

	// Cleanup job coordination methods for serverless/multi-instance environments
	// TryAcquireCleanupLock attempts to acquire the cleanup lock and checks if cleanup should run.
	// Returns true if the lock was acquired and enough time has passed since last run.
	TryAcquireCleanupLock(ctx context.Context, minInterval time.Duration) (bool, domain.Error)
	// ReleaseCleanupLock releases the cleanup lock (e.g., on error)
	ReleaseCleanupLock(ctx context.Context) domain.Error
	// UpdateCleanupLastRun updates the last run timestamp and releases the lock
	UpdateCleanupLastRun(ctx context.Context) domain.Error
}
