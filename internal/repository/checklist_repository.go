package repository

import (
	"context"
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/repository/connection"
	"com.raunlo.checklist/internal/repository/dbo"
	"com.raunlo.checklist/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/raunlo/pgx-with-automapper/mapper"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type checklistRepository struct {
	connection pool.Conn
}

func (repository *checklistRepository) UpdateChecklist(checklist domain.Checklist) (domain.Checklist, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		query := `UPDATE checklist
				  SET NAME = @checklist_name
				  WHERE ID = @checklist_id`
		succeeded, err := tx.Exec(context.Background(), query, pgx.NamedArgs{
			"checklist_name": checklist.Name,
			"checklist_id":   checklist.Id,
		})

		return succeeded.RowsAffected() == 1, err
	}
	res, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Query:      queryFunc,
		TxOptions:  pgx.TxOptions{IsoLevel: pgx.Serializable},
		Connection: repository.connection,
	})

	if err != nil {
		return domain.Checklist{}, domain.Wrap(err, "Could not save checklist", 500)
	} else if !res {
		return domain.Checklist{}, domain.NewError(
			fmt.Sprintf("Could not update checklist(id=%d) because it was non-existant", checklist.Id),
			500)
	} else {
		return checklist, nil
	}
}

func (repository *checklistRepository) SaveChecklist(checklist domain.Checklist) (domain.Checklist, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (domain.Checklist, error) {
		query := `INSERT INTO checklist(ID, NAME) 
				  VALUES (nextval('checklist_id_sequence'), @checklist_name) RETURNING ID`
		row := tx.QueryRow(context.Background(), query, pgx.NamedArgs{
			"checklist_name": checklist.Name,
		})

		err := row.Scan(&checklist.Id)
		if err == nil {
			err = repository.createPhantomChecklistItem(tx, checklist.Id)
		}
		return checklist, err

	}
	res, err := connection.RunInTransaction(connection.TransactionProps[domain.Checklist]{
		Query:      queryFunc,
		Connection: repository.connection,
		TxOptions:  pgx.TxOptions{IsoLevel: pgx.Serializable},
	})

	if err != nil {
		return domain.Checklist{}, domain.Wrap(err, "Could not save checklist", 500)
	} else {
		return res, nil
	}
}

func (repository *checklistRepository) FindChecklistById(id uint) (*domain.Checklist, domain.Error) {
	const query = "SELECT id, name FROM checklist where ID = @checklist_id"
	var checklistDbo dbo.ChecklistDbo
	err := repository.connection.QueryOne(context.Background(), query, &checklistDbo, pgx.NamedArgs{
		"checklist_id": id,
	})

	if errors.Is(err, mapper.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, domain.Wrap(err, fmt.Sprintf("Failed to find checklist(id=%d)", id), 500)
	}

	return util.AnyPointer(dbo.MapChecklistDboToDomain(checklistDbo)), nil
}

func (repository *checklistRepository) DeleteChecklistById(id uint) domain.Error {
	runQueryFunction := func(tx pool.TransactionWrapper) (bool, error) {
		sqlQueryNamedArgs := pgx.NamedArgs{
			"checklist_id": id,
		}
		result, err := tx.Exec(context.Background(), "DELETE FROM checklist where ID = @checklist_id", sqlQueryNamedArgs)
		return result.RowsAffected() == 1, err
	}

	res, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Query:      runQueryFunction,
		Connection: repository.connection,
		TxOptions:  pgx.TxOptions{IsoLevel: pgx.Serializable},
	})

	if err != nil {
		return domain.Wrap(err, "Failed to find all checklists", 500)
	} else if !res {
		return domain.NewError(fmt.Sprintf("Could not delete checklist, because it did not exist with id: %d", id), 500)
	} else {
		return nil
	}
}

func (repository *checklistRepository) FindAllChecklists() ([]domain.Checklist, domain.Error) {
	checklistDbos := make([]dbo.ChecklistDbo, 0)
	err := repository.connection.QueryList(context.TODO(), "SELECT ID, NAME FROM CHECKLIST", &checklistDbos, pgx.NamedArgs{})

	if err != nil {
		return nil, domain.Wrap(err, "Failed to find all checklists", 500)
	} else {
		var checklists []domain.Checklist
		for _, checklistDbo := range checklistDbos {
			checklists = append(checklists, dbo.MapChecklistDboToDomain(checklistDbo))
		}
		return checklists, nil
	}
}

func (repository *checklistRepository) createPhantomChecklistItem(tx pool.TransactionWrapper, checklistId uint) error {

	sql := `INSERT INTO CHECKLIST_ITEM(CHECKLIST_ITEM_ID, CHECKLIST_ID, CHECKLIST_ITEM_NAME, CHECKLIST_ITEM_COMPLETED, IS_PHANTOM, NEXT_ITEM_ID, PREV_ITEM_ID)
				 VALUES(nextval('checklist_item_id_sequence'), @checklistId, @phantomItemName, false, true, null, null)`

	_, err := tx.Exec(context.Background(), sql, pgx.NamedArgs{
		"checklistId":     checklistId,
		"phantomItemName": "__phantom__",
	})
	return err
}
