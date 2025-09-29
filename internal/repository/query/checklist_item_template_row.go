package query

import (
	"com.raunlo.checklist/internal/core/domain"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type PersistChecklistItemTemplateRowQueryFunction struct {
	checklistItemTemplateId   uint
	checklistItemTemplateRows []domain.ChecklistItemTemplateRow
}

func (p *PersistChecklistItemTemplateRowQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) ([]domain.ChecklistItemTemplateRow, error) {
	return func(tx pool.TransactionWrapper) ([]domain.ChecklistItemTemplateRow, error) {
		// TODO implement me
		panic("implement me")
	}
}

type UpdateChecklistItemTemplateRowsQueryFunction struct {
	checklistItemTemplateId   uint
	checklistItemTemplateRows []domain.ChecklistItemTemplateRow
}

func (u *UpdateChecklistItemTemplateRowsQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		// TODO implement me
		panic("implement me")
	}
}
