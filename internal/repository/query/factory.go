package query

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/repository/dbo"
	"github.com/raunlo/pgx-with-automapper/pool"
)

// TransactionalQuery For query structs that use Transactional queries
type TransactionalQuery[K any] interface {
	GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (K, error)
}

// Query For Queries that dont use transactions
type Query[K any] interface {
	GetQueryFunction(ctx context.Context) func(connection pool.Conn) (K, error)
}

func NewPersistChecklistItemQueryFunction(checklistId uint, checklistItem domain.ChecklistItem) TransactionalQuery[domain.ChecklistItem] {
	return &PersistChecklistItemQueryFunction{
		checklistItem: checklistItem,
		checklistId:   checklistId,
	}
}

func NewPersistChecklistItemRowsQueryFunction(checklistItemId uint, checklistItemRows []domain.ChecklistItemRow) TransactionalQuery[[]domain.ChecklistItemRow] {
	return &PersistChecklistItemRowQueryFunction{
		checklistItemRows: checklistItemRows,
		checklistItemId:   checklistItemId,
	}
}

func NewGetAllChecklistItemsWithRowsQueryFunction(checklistId uint, completed *bool, sortOrder domain.SortOrder) Query[[]dbo.ChecklistItemDbo] {
	return &GetAllChecklistItemsQueryFunction{
		checklistId: checklistId,
		completed:   completed,
		sortOrder:   sortOrder,
	}
}

func NewDeleteChecklistItemByIdQueryFunction(checklistId uint, checklistItemId uint) TransactionalQuery[bool] {
	return &DeleteChecklistItemQueryFunction{
		checklistId:     checklistId,
		checklistItemId: checklistItemId,
	}
}

func NewFindChecklistItemByIdQueryFunction(checklistId uint, checklistItemId uint) Query[*dbo.ChecklistItemDbo] {
	return &FindChecklistItemById{
		checklistId:     checklistId,
		checklistItemId: checklistItemId,
	}
}

func NewUpdateChecklistItemQueryFunction(checklistId uint, checklistItem domain.ChecklistItem) TransactionalQuery[bool] {
	return &UpdateChecklistItemFunction{
		checklistId:   checklistId,
		checklistItem: checklistItem,
	}
}

func NewUpdateChecklistItemRowsQueryFunction(checklistItemId uint, rows []domain.ChecklistItemRow) TransactionalQuery[bool] {
	return &UpdateChecklistItemRowsQueryFunction{
		checklistItemId:   checklistItemId,
		checklistItemRows: rows,
	}
}

func NewDeleteChecklistItemRowByIdQueryFunction(checklistId uint, checklistItemId uint, rowId uint) TransactionalQuery[bool] {
	return &DeleteChecklistItemRowQueryFunction{
		checklistId:     checklistId,
		checklistItemId: checklistItemId,
		rowId:           rowId,
	}
}

func NewUpdateChecklistItemTemplateRowsQueryFunction(checklistItemTemplateId uint, rows []domain.ChecklistItemTemplateRow) TransactionalQuery[bool] {
	return &UpdateChecklistItemTemplateRowsQueryFunction{
		checklistItemTemplateId:   checklistItemTemplateId,
		checklistItemTemplateRows: rows,
	}
}

func NewUpdateChecklistItemTemplateQueryFunction(checklistItemTemplate domain.ChecklistItemTemplate) TransactionalQuery[bool] {
	return &UpdateChecklistItemTemplateQueryFunction{
		checklistItemTemplate: checklistItemTemplate,
	}
}

func NewPersistChecklistItemTemplateRowsQueryFunction(checklistItemTemplateId uint, rows []domain.ChecklistItemTemplateRow) TransactionalQuery[[]domain.ChecklistItemTemplateRow] {
	return &PersistChecklistItemTemplateRowQueryFunction{
		checklistItemTemplateId:   checklistItemTemplateId,
		checklistItemTemplateRows: rows,
	}
}

func NewPersistChecklistItemTemplateQueryFunction(checklistTemplate domain.ChecklistItemTemplate) TransactionalQuery[domain.ChecklistItemTemplate] {
	return &PersistChecklistItemTemplateQueryFunction{
		checklistItemTemplate: checklistTemplate,
	}
}

func NewDeleteChecklistItemTemplateByIdQueryFunction(checklistItemTemplateId uint) TransactionalQuery[bool] {
	return &DeleteChecklistItemTemplateByIdQueryFunction{
		checklistItemTemplateId: checklistItemTemplateId,
	}
}

func NewFindChecklistItemTemplateByIdQueryFunction(checklistItemTemplateId uint) Query[*domain.ChecklistItemTemplate] {
	return &FindChecklistItemTemplateByIdQueryFunction{
		checklistItemTemplateId: checklistItemTemplateId,
	}
}

func NewGetAllChecklistItemTemplatesQueryFunction() Query[[]domain.ChecklistItemTemplate] {
	return &GetAllChecklistItemTemplatesQueryFunction{}
}

func NewChangeChecklistItemOrderQueryFunction(changeOrderRequest domain.ChangeOrderRequest) TransactionalQuery[bool] {
	return &ChangeChecklistItemOrderQueryFunction{
		checklistId:     changeOrderRequest.ChecklistId,
		checklistItemId: changeOrderRequest.ChecklistItemId,
		newOrderNumber:  changeOrderRequest.NewOrderNumber,
		sortOrder:       changeOrderRequest.SortOrder,
	}
}

func NewToggleCompletionQueryFunction(checklistId uint, checklistItemId uint, completed bool) TransactionalQuery[domain.ChecklistItem] {
	return &toggleCompletionQueryFunction{
		checklistId:     checklistId,
		checklistItemId: checklistItemId,
		completed:       completed,
	}
}
