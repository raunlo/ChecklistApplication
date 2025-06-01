package repository

import (
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/repository/connection"
	"com.raunlo.checklist/internal/repository/query"
	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type IChecklistItemTemplateRepository interface {
	SaveChecklistTemplate(checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error)
	GetAllChecklistTemplates() ([]domain.ChecklistItemTemplate, domain.Error)
	UpdateChecklistTemplate(checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error)
	DeleteChecklistTemplateById(id uint) domain.Error
	FindChecklistTemplateById(id uint) (*domain.ChecklistItemTemplate, domain.Error)
}

type checklistItemTemplateRepository struct {
	connection pool.Conn
}

func (repository *checklistItemTemplateRepository) SaveChecklistTemplate(checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error) {
	queryFunction := func(tx pool.TransactionWrapper) (domain.ChecklistItemTemplate, error) {
		savedChecklistItemTemplate, err := query.NewPersistChecklistItemTemplateQueryFunction(checklistTemplate).
			GetTransactionalQueryFunction()(tx)
		if err != nil {
			return domain.ChecklistItemTemplate{}, err
		}
		rows, err := query.NewPersistChecklistItemTemplateRowsQueryFunction(savedChecklistItemTemplate.Id, checklistTemplate.Rows).
			GetTransactionalQueryFunction()(tx)
		savedChecklistItemTemplate.Rows = rows
		return savedChecklistItemTemplate, err
	}

	res, err := connection.RunInTransaction(connection.TransactionProps[domain.ChecklistItemTemplate]{
		TxOptions:  pgx.TxOptions{IsoLevel: pgx.Serializable},
		Query:      queryFunction,
		Connection: repository.connection,
	})

	if err != nil {
		return domain.ChecklistItemTemplate{}, domain.Wrap(err, "Failed to persist checklistItemTemplate", 500)
	}
	return res, nil
}

func (repository *checklistItemTemplateRepository) GetAllChecklistTemplates() ([]domain.ChecklistItemTemplate, domain.Error) {
	res, err := query.NewGetAllChecklistItemTemplatesQueryFunction().GetQueryFunction()(repository.connection)

	if err != nil {
		return nil, domain.Wrap(err, "Failed to find all checklistItemTemplates", 500)
	}
	return res, nil
}

func (repository *checklistItemTemplateRepository) UpdateChecklistTemplate(checklistTemplate domain.ChecklistItemTemplate) (domain.ChecklistItemTemplate, domain.Error) {
	queryFunction := func(tx pool.TransactionWrapper) (bool, error) {
		_, err := query.NewUpdateChecklistItemTemplateQueryFunction(checklistTemplate).GetTransactionalQueryFunction()(tx)
		if err != nil {
			return false, err
		}
		ok, err := query.NewUpdateChecklistItemTemplateRowsQueryFunction(checklistTemplate.Id, checklistTemplate.Rows).
			GetTransactionalQueryFunction()(tx)
		return ok, err
	}

	ok, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		TxOptions:  pgx.TxOptions{IsoLevel: pgx.Serializable},
		Query:      queryFunction,
		Connection: repository.connection,
	})

	if err != nil {
		return domain.ChecklistItemTemplate{}, domain.Wrap(err,
			fmt.Sprintf("Failed to update checklistItemTemplate(id=%d)", checklistTemplate.Id),
			500)
	} else if !ok {
		return domain.ChecklistItemTemplate{}, domain.NewError(
			fmt.Sprintf("ChecklistItemTemplate(id=%d) does not exists", checklistTemplate.Id),
			404)
	}
	return checklistTemplate, nil
}

func (repository *checklistItemTemplateRepository) DeleteChecklistTemplateById(id uint) domain.Error {
	res, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		TxOptions:  pgx.TxOptions{IsoLevel: pgx.Serializable},
		Query:      query.NewDeleteChecklistItemTemplateByIdQueryFunction(id).GetTransactionalQueryFunction(),
		Connection: repository.connection,
	})

	if err != nil {
		return domain.Wrap(err,
			fmt.Sprintf("Failed to delete checklistItemTemplate(id=%d)", id),
			500)
	} else if !res {
		return domain.NewError(fmt.Sprintf("ChecklistItemTemplate(id=%d) does not exists", id), 404)
	}
	return nil

}

func (repository *checklistItemTemplateRepository) FindChecklistTemplateById(id uint) (*domain.ChecklistItemTemplate, domain.Error) {
	res, err := query.NewFindChecklistItemTemplateByIdQueryFunction(id).GetQueryFunction()(repository.connection)
	if err != nil {
		return nil, domain.Wrap(err, "Failed to find checklistITemTemplate", 500)
	}
	return res, nil
}
