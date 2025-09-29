package query

import (
	"com.raunlo.checklist/internal/core/domain"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type PersistChecklistItemTemplateQueryFunction struct {
	checklistItemTemplate domain.ChecklistItemTemplate
}

func (p *PersistChecklistItemTemplateQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (domain.ChecklistItemTemplate, error) {
	return func(tx pool.TransactionWrapper) (domain.ChecklistItemTemplate, error) {
		// TODO implement me
		panic("implement me")
	}
}

type UpdateChecklistItemTemplateQueryFunction struct {
	checklistItemTemplate domain.ChecklistItemTemplate
}

func (u *UpdateChecklistItemTemplateQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		// TODO implement me
		panic("implement me")
	}
}

type DeleteChecklistItemTemplateByIdQueryFunction struct {
	checklistItemTemplateId uint
}

func (d *DeleteChecklistItemTemplateByIdQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		// TODO implement me
		panic("implement me")
	}
}

type FindChecklistItemTemplateByIdQueryFunction struct {
	checklistItemTemplateId uint
}

func (f FindChecklistItemTemplateByIdQueryFunction) GetQueryFunction() func(connection pool.Conn) (*domain.ChecklistItemTemplate, error) {
	return func(connection pool.Conn) (*domain.ChecklistItemTemplate, error) {
		// TODO implement me
		panic("implement me")
	}
}

type GetAllChecklistItemTemplatesQueryFunction struct{}

func (g *GetAllChecklistItemTemplatesQueryFunction) GetQueryFunction() func(connection pool.Conn) ([]domain.ChecklistItemTemplate, error) {
	return func(connection pool.Conn) ([]domain.ChecklistItemTemplate, error) {
		// TODO implement me
		panic("implement me")
	}
}
