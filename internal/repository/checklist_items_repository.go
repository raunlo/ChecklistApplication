package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/repository/connection"
	"com.raunlo.checklist/internal/repository/dbo"
	"com.raunlo.checklist/internal/repository/query"
	"github.com/raunlo/pgx-with-automapper/mapper"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type checklistItemRepository struct {
	conn pool.Conn
}

func (r *checklistItemRepository) FindChecklistItemById(ctx context.Context, checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
	result, err := query.NewFindChecklistItemByIdQueryFunction(checklistId, id).GetQueryFunction(ctx)(r.conn)

	if errors.Is(err, mapper.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, domain.Wrap(err,
			fmt.Sprintf("Error occured on finding checklistItem(checklistId=%d, checklistItemId=%d)", checklistId, id),
			500)
	}
	return new(dbo.MapChecklistItemDboToDomain(*result)), nil
}

func (r *checklistItemRepository) UpdateChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	queryFunction := func(tx pool.TransactionWrapper) (bool, error) {
		_, err := query.NewUpdateChecklistItemQueryFunction(checklistId, checklistItem).GetTransactionalQueryFunction()(tx)
		if err != nil {
			return false, err
		}
		ok, err := query.NewUpdateChecklistItemRowsQueryFunction(checklistItem.Id, checklistItem.Rows).GetTransactionalQueryFunction()(tx)

		return ok, err
	}

	res, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		TxOptions:  connection.TxReadCommitted, // Simple single-item update
		Connection: r.conn,
		Query:      queryFunction,
	})

	if err != nil {
		return domain.ChecklistItem{}, domain.Wrap(err, "Failed to update checklistItem", 500)
	} else if !res {
		return domain.ChecklistItem{}, domain.NewError("ChecklistItem was not found", 404)
	}

	return checklistItem, nil
}

func (r *checklistItemRepository) SaveChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	queryFunction := func(tx pool.TransactionWrapper) (domain.ChecklistItem, error) {
		savedChecklistItem, err := query.NewPersistChecklistItemQueryFunction(checklistId, checklistItem).GetTransactionalQueryFunction()(tx)
		if err == nil {
			var rows []domain.ChecklistItemRow
			rows, err = query.NewPersistChecklistItemRowsQueryFunction(savedChecklistItem.Id, checklistItem.Rows).GetTransactionalQueryFunction()(tx)
			savedChecklistItem.Rows = rows
		}
		return savedChecklistItem, err
	}

	res, err := connection.RunInTransaction(connection.TransactionProps[domain.ChecklistItem]{
		Ctx:        ctx,
		TxOptions:  connection.TxReadCommitted, // Simple insert operation
		Query:      queryFunction,
		Connection: r.conn,
	})
	if err != nil {
		return domain.ChecklistItem{}, domain.Wrap(err, "Could not save checklistItem", 500)
	}

	return res, nil
}

func (r *checklistItemRepository) SaveChecklistItemRow(ctx context.Context, checklistId uint, checklistItemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error) {
	if _, err := query.NewFindChecklistItemByIdQueryFunction(checklistId, checklistItemId).GetQueryFunction(ctx)(r.conn); errors.Is(err, mapper.ErrNoRows) {
		return domain.ChecklistItemRow{}, domain.NewError("ChecklistItem was not found", 404)
	} else if err != nil {
		return domain.ChecklistItemRow{}, domain.Wrap(err,
			fmt.Sprintf("Error occured on finding checklistItem(checklistId=%d, checklistItemId=%d)", checklistId, checklistItemId),
			500)
	}

	queryFunction := func(tx pool.TransactionWrapper) ([]domain.ChecklistItemRow, error) {
		return query.NewPersistChecklistItemRowsQueryFunction(checklistItemId, []domain.ChecklistItemRow{row}).GetTransactionalQueryFunction()(tx)
	}

	res, err := connection.RunInTransaction(connection.TransactionProps[[]domain.ChecklistItemRow]{
		Ctx:        ctx,
		TxOptions:  connection.TxReadCommitted, // Simple row insert
		Connection: r.conn,
		Query:      queryFunction,
	})

	if err != nil {
		return domain.ChecklistItemRow{}, domain.Wrap(err, "Could not save checklistItemRow", 500)
	} else if len(res) == 0 {
		return domain.ChecklistItemRow{}, domain.NewError("Failed to save checklistItemRow", 500)
	}

	return res[0], nil
}

func (r *checklistItemRepository) DeleteChecklistItemById(ctx context.Context, checklistId uint, id uint) domain.Error {
	result, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		TxOptions:  connection.TxSerializable, // Serializable for SELECT...FOR UPDATE locking
		Connection: r.conn,
		Query:      query.NewDeleteChecklistItemByIdQueryFunction(checklistId, id).GetTransactionalQueryFunction(),
	})

	if err != nil {
		return domain.Wrap(err, "Could not delete checklistItem due an error", 500)
	} else if !result {
		return domain.NewError("Failed to delete checklistItem", 404)
	}
	return nil
}

func (r *checklistItemRepository) DeleteChecklistItemRowAndAutoComplete(ctx context.Context, checklistId uint, checklistItemId uint, rowId uint) (domain.ChecklistItemRowDeletionResult, domain.Error) {
	result, err := connection.RunInTransaction(connection.TransactionProps[domain.ChecklistItemRowDeletionResult]{
		Ctx:        ctx,
		TxOptions:  connection.TxSerializable, // Multi-row atomic: delete + auto-complete check
		Connection: r.conn,
		Query:      query.NewDeleteChecklistItemRowAndAutoCompleteQueryFunction(checklistId, checklistItemId, rowId).GetTransactionalQueryFunction(),
	})

	if err != nil {
		return domain.ChecklistItemRowDeletionResult{Success: false, ItemAutoCompleted: false}, domain.Wrap(err, "Could not delete checklistItemRow due an error", 500)
	} else if !result.Success {
		return domain.ChecklistItemRowDeletionResult{Success: false, ItemAutoCompleted: false}, domain.NewError("Failed to delete checklistItemRow", 404)
	}
	return result, nil
}

func (r *checklistItemRepository) FindAllChecklistItems(ctx context.Context, checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error) {
	dbos, err := query.NewGetAllChecklistItemsWithRowsQueryFunction(checklistId, completed, sortOrder).
		GetQueryFunction(ctx)(r.conn)
	if err != nil {
		return nil, domain.Wrap(err, "Failed to query checklistItems", 500)
	}
	var items []domain.ChecklistItem
	for _, item := range dbos {
		items = append(items, dbo.MapChecklistItemDboToDomain(item))
	}
	return items, nil
}

func (r *checklistItemRepository) ChangeChecklistItemOrder(ctx context.Context, request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error) {
	response, err := connection.RunInTransaction(connection.TransactionProps[domain.ChangeOrderResponse]{
		Ctx:        ctx,
		Connection: r.conn,
		Query:      query.NewChangeChecklistItemOrderQueryFunction(request).GetTransactionalQueryFunction(),
		TxOptions:  connection.TxSerializable, // Ordering requires strict consistency
	})

	if err != nil {
		return domain.ChangeOrderResponse{}, domain.Wrap(err, "Error happened during changing checklist item order number", 500)
	}
	return response, nil
}

func (r *checklistItemRepository) ToggleItemCompleted(ctx context.Context, checklistId uint, checklistItemId uint, completed bool) (domain.ChecklistItem, domain.Error) {
	queryFunction := query.NewToggleCompletionQueryFunction(checklistId, checklistItemId, completed)

	res, err := connection.RunInTransaction(connection.TransactionProps[domain.ChecklistItem]{
		Ctx:        ctx,
		Query:      queryFunction.GetTransactionalQueryFunction(),
		TxOptions:  connection.TxReadCommitted, // Simple single-row toggle
		Connection: r.conn,
	})
	if err != nil {
		return domain.ChecklistItem{}, domain.Wrap(err, "Failed to mark item as completed", 500)
	}
	return res, nil
}

func (r *checklistItemRepository) RebalancePositions(ctx context.Context, checklistId uint) domain.Error {
	_, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Connection: r.conn,
		Query:      query.NewRebalancePositionsQueryFunction(checklistId).GetTransactionalQueryFunction(),
		TxOptions:  connection.TxSerializable, // Multi-row rebalance requires strict consistency
	})

	if err != nil {
		return domain.Wrap(err, "Error happened during rebalancing positions", 500)
	}
	return nil
}

func (r *checklistItemRepository) RestoreChecklistItem(ctx context.Context, checklistId uint, itemId uint) (domain.ChecklistItem, domain.Error) {
	result, err := connection.RunInTransaction(connection.TransactionProps[dbo.ChecklistItemDbo]{
		Ctx:        ctx,
		TxOptions:  connection.TxSerializable, // Serializable for SELECT...FOR UPDATE locking
		Connection: r.conn,
		Query:      query.NewRestoreChecklistItemQueryFunction(checklistId, itemId).GetTransactionalQueryFunction(),
	})

	if err != nil {
		return domain.ChecklistItem{}, domain.Wrap(err, "Could not restore checklistItem", 500)
	}

	return dbo.MapChecklistItemDboToDomain(result), nil
}

func (r *checklistItemRepository) PurgeSoftDeletedItems(ctx context.Context, retentionPeriod time.Duration) (int64, domain.Error) {
	retentionHours := int(retentionPeriod.Hours())

	result, err := connection.RunInTransaction(connection.TransactionProps[int64]{
		Ctx:        ctx,
		TxOptions:  connection.TxReadCommitted,
		Connection: r.conn,
		Query:      query.NewPurgeSoftDeletedItemsQueryFunction(retentionHours).GetTransactionalQueryFunction(),
	})

	if err != nil {
		return 0, domain.Wrap(err, "Could not purge soft-deleted items", 500)
	}

	return result, nil
}

func (r *checklistItemRepository) TryAcquireCleanupLock(ctx context.Context, minInterval time.Duration) (bool, domain.Error) {
	result, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		TxOptions:  connection.TxSerializable, // Serializable for locking
		Connection: r.conn,
		Query:      query.NewTryAcquireCleanupLockQueryFunction(minInterval).GetTransactionalQueryFunction(),
	})
	if err != nil {
		return false, domain.Wrap(err, "Could not acquire cleanup lock", 500)
	}
	return result, nil
}

func (r *checklistItemRepository) ReleaseCleanupLock(ctx context.Context) domain.Error {
	_, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		TxOptions:  connection.TxReadCommitted, // Simple UPDATE
		Connection: r.conn,
		Query:      query.NewReleaseCleanupLockQueryFunction().GetTransactionalQueryFunction(),
	})
	if err != nil {
		return domain.Wrap(err, "Could not release cleanup lock", 500)
	}
	return nil
}

func (r *checklistItemRepository) UpdateCleanupLastRun(ctx context.Context) domain.Error {
	_, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		TxOptions:  connection.TxReadCommitted, // Simple UPDATE
		Connection: r.conn,
		Query:      query.NewUpdateCleanupLastRunQueryFunction().GetTransactionalQueryFunction(),
	})
	if err != nil {
		return domain.Wrap(err, "Could not update cleanup last run", 500)
	}
	return nil
}
