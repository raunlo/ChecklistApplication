package repository

import (
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/repository/connection"
	"com.raunlo.checklist/internal/repository/dbo"
	"com.raunlo.checklist/internal/repository/query"
	"com.raunlo.checklist/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type IChecklistItemsRepository interface {
	UpdateChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	SaveChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	FindChecklistItemById(checklistId uint, id uint) (*domain.ChecklistItem, domain.Error)
	DeleteChecklistItemById(checklistId uint, id uint) domain.Error
	FindAllChecklistItems(checklistId uint, completed bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error)
	ChangeChecklistItemOrder(request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error)
}

type checklistItemRepository struct {
	conn pool.Conn
}

func (r *checklistItemRepository) FindChecklistItemById(checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
	result, err := query.NewFindChecklistItemByIdQueryFunction(checklistId, id).GetQueryFunction()(r.conn)
	if err != nil {
		return nil, domain.Wrap(err,
			fmt.Sprintf("Error occured on finding checklistItem(checklistId=%d, checklistItemId=%d)", checklistId, id),
			500)
	} else if result == nil {
		return nil, domain.NewError("ChecklistItem was not found", 404)
	}
	return util.AnyPointer(dbo.MapChecklistItemDboToDomain(*result)), nil
}

func (r *checklistItemRepository) UpdateChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	queryFunction := func(tx pool.TransactionWrapper) (bool, error) {
		_, err := query.NewUpdateChecklistItemQueryFunction(checklistId, checklistItem).GetTransactionalQueryFunction()(tx)
		if err != nil {
			return false, nil
		}
		ok, err := query.NewUpdateChecklistItemRowsQueryFunction(checklistItem.Id, checklistItem.Rows).GetTransactionalQueryFunction()(tx)

		return ok, err
	}

	res, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		TxOptions:  pgx.TxOptions{IsoLevel: pgx.Serializable},
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

func (r *checklistItemRepository) SaveChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
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
		TxOptions:  pgx.TxOptions{IsoLevel: pgx.Serializable},
		Query:      queryFunction,
		Connection: r.conn,
	})

	if err != nil {
		return domain.ChecklistItem{}, domain.Wrap(err, "Could not save checklistItem", 500)
	}

	return res, nil
}

func (r *checklistItemRepository) DeleteChecklistItemById(checklistId uint, id uint) domain.Error {
	result, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		TxOptions:  pgx.TxOptions{IsoLevel: pgx.Serializable},
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

func (r *checklistItemRepository) FindAllChecklistItems(checklistId uint, completed bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error) {
	dbos, err := query.NewGetAllChecklistItemsWithRowsQueryFunction(checklistId, &completed, sortOrder).
		GetQueryFunction()(r.conn)

	if err != nil {
		return nil, domain.Wrap(err, "Failed to query checklistItems", 500)
	}
	var items []domain.ChecklistItem
	for _, item := range dbos {
		items = append(items, dbo.MapChecklistItemDboToDomain(item))
	}
	return items, nil
}

func (r *checklistItemRepository) ChangeChecklistItemOrder(request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error) {
	ok, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Connection: r.conn,
		Query:      query.NewChangeChecklistItemOrderQueryFunction(request).GetTransactionalQueryFunction(),
		TxOptions:  pgx.TxOptions{IsoLevel: pgx.Serializable},
	})

	if err != nil {
		return domain.ChangeOrderResponse{}, domain.Wrap(err, "Error happened during changing checklist item order number", 500)
	} else if !ok {
		return domain.ChangeOrderResponse{}, domain.NewError("Failed to change checklist item order", 400)
	}
	return domain.ChangeOrderResponse{
		OrderNumber:     request.NewOrderNumber,
		ChecklistItemId: request.ChecklistItemId,
		ChecklistId:     request.ChecklistId,
	}, nil
}
