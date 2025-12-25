package repository

import (
	"context"
	"fmt"
	"log"

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

func (repository *checklistRepository) UpdateChecklist(ctx context.Context, checklist domain.Checklist) (domain.Checklist, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		query := `UPDATE checklist
				  SET NAME = @checklist_name
				  WHERE ID = @checklist_id`
		succeeded, err := tx.Exec(ctx, query, pgx.NamedArgs{
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

func (repository *checklistRepository) SaveChecklist(ctx context.Context, checklist domain.Checklist) (domain.Checklist, domain.Error) {
	owner, userIdError := domain.GetUserIdFromContext(ctx)
	if userIdError != nil {
		return domain.Checklist{}, userIdError
	}

	queryFunc := func(tx pool.TransactionWrapper) (domain.Checklist, error) {
		query := `INSERT INTO checklist(ID, NAME, OWNER) 
				  VALUES (nextval('checklist_id_sequence'), @checklist_name, @owner) RETURNING ID`
		row := tx.QueryRow(ctx, query, pgx.NamedArgs{
			"checklist_name": checklist.Name,
			"owner":          owner,
		})

		err := row.Scan(&checklist.Id)
		if err == nil {
			err = repository.createPhantomChecklistItem(tx, checklist.Id)
		}
		checklist.Owner = owner
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

func (repository *checklistRepository) FindChecklistById(ctx context.Context, id uint) (*domain.Checklist, domain.Error) {
	const query = "SELECT id, name FROM checklist where ID = @checklist_id"
	var checklistDbo dbo.ChecklistDbo
	err := repository.connection.QueryOne(ctx, query, &checklistDbo, pgx.NamedArgs{
		"checklist_id": id,
	})

	if errors.Is(err, mapper.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, domain.Wrap(err, fmt.Sprintf("Failed to find checklist(id=%d)", id), 500)
	}

	return util.AnyPointer(dbo.MapChecklistDboToDomain(checklistDbo)), nil
}

func (repository *checklistRepository) DeleteChecklistById(ctx context.Context, id uint) domain.Error {
	runQueryFunction := func(tx pool.TransactionWrapper) (bool, error) {
		sqlQueryNamedArgs := pgx.NamedArgs{
			"checklist_id": id,
		}
		result, err := tx.Exec(ctx, "DELETE FROM checklist where ID = @checklist_id", sqlQueryNamedArgs)
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

func (repository *checklistRepository) FindAllChecklists(ctx context.Context) ([]domain.Checklist, domain.Error) {
	query := `
		SELECT 
			c.ID as id,
			c.NAME as name,
			c.OWNER as owner,
			COALESCE(COUNT(ci.checklist_item_id) FILTER (WHERE ci.is_phantom = false), 0) as total_items,
			COALESCE(COUNT(ci.checklist_item_id) FILTER (WHERE ci.is_phantom = false AND ci.checklist_item_completed = true), 0) as completed_items,
			ARRAY_REMOVE(ARRAY_AGG(DISTINCT cs.SHARED_WITH_USER_ID), NULL) as shared_with
		FROM CHECKLIST c
		LEFT JOIN CHECKLIST_SHARE cs ON c.ID = cs.CHECKLIST_ID
		LEFT JOIN CHECKLIST_ITEM ci ON c.ID = ci.CHECKLIST_ID
		WHERE c.OWNER = @user_id OR cs.SHARED_WITH_USER_ID = @user_id
		GROUP BY c.ID, c.NAME, c.OWNER
		ORDER BY c.ID DESC
	`

	userId, ok := ctx.Value(domain.UserIdContextKey).(string)
	if !ok {
		return nil, domain.NewError("User ID not found in context", 401)
	}

	rows, err := repository.connection.Query(ctx, query, pgx.NamedArgs{
		"user_id": userId,
	})
	if err != nil {
		return nil, domain.Wrap(err, "Failed to query checklists", 500)
	}
	defer rows.Close()

	var checklists []domain.Checklist
	for rows.Next() {
		var id uint
		var name string
		var owner string
		var totalItems int64
		var completedItems int64
		var sharedWith []string

		err := rows.Scan(&id, &name, &owner, &totalItems, &completedItems, &sharedWith)
		if err != nil {
			return nil, domain.Wrap(err, "Failed to scan checklist row", 500)
		}

		checklist := domain.Checklist{
			Id:         id,
			Name:       name,
			Owner:      owner,
			SharedWith: sharedWith,
		}

		// Create ChecklistItems array with stats info
		// We store count as phantom items to pass stats through domain layer
		totalItemsInt := int(totalItems)
		completedItemsInt := int(completedItems)

		checklist.ChecklistItems = make([]domain.ChecklistItem, totalItemsInt)
		for i := 0; i < completedItemsInt; i++ {
			checklist.ChecklistItems[i].Completed = true
		}

		checklists = append(checklists, checklist)
	}

	if rows.Err() != nil {
		return nil, domain.Wrap(rows.Err(), "Error iterating checklist rows", 500)
	}

	return checklists, nil
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

func (repository *checklistRepository) CheckUserHasAccessToChecklist(ctx context.Context, checklistId uint, userId string) (bool, domain.Error) {
	query := `
		SELECT
			(c.owner = @user_id) AS is_owner,
			cs.PERMISSION_LEVEL
		FROM checklist c
		LEFT JOIN checklist_share cs ON cs.checklist_id = c.id
		WHERE c.id = @checklist_id AND (c.owner = @user_id OR cs.shared_with_user_id = @user_id) 
		LIMIT 1
		`
	var isOwner bool
	var shareLevel *string
	err := repository.connection.QueryRow(ctx, query, pgx.NamedArgs{
		"checklist_id": checklistId,
		"user_id":      userId,
	}).Scan(&isOwner, &shareLevel)

	// Check for errors first. ErrNoRows means no access (not an error case).
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No matching row means user has no access
			return false, nil
		}
		return false, domain.Wrap(err, "Failed to check user access to checklist", 500)
	}

	hasAccess := isOwner || (shareLevel != nil)

	// Optional: emit info so caller can understand whether access is owner or shared and the level.
	// This keeps the function signature unchanged while surfacing the details to logs.
	if shareLevel != nil {
		if isOwner {
			log.Printf("User(id=%s) is owner of checklist %d (share entry also present: level=%s)", userId, checklistId, *shareLevel)
		} else {
			log.Printf("User(id=%s) has shared access to checklist %d with level=%s", userId, checklistId, *shareLevel)
		}
	} else if isOwner {
		log.Printf("User(id=%s) has owner access to checklist %d", userId, checklistId)
	}

	return hasAccess, nil
}

func (repository *checklistRepository) CheckUserIsOwner(ctx context.Context, checklistId uint, userId string) (bool, domain.Error) {
	query := `SELECT (c.owner = @user_id) AS is_owner
			  FROM checklist c
			  WHERE c.id = @checklist_id`

	var isOwner bool
	err := repository.connection.QueryRow(ctx, query, pgx.NamedArgs{
		"checklist_id": checklistId,
		"user_id":      userId,
	}).Scan(&isOwner)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Checklist doesn't exist
			return false, nil
		}
		return false, domain.Wrap(err, "Failed to check checklist ownership", 500)
	}

	return isOwner, nil
}

func (repository *checklistRepository) CreateChecklistShare(ctx context.Context, checklistId uint, sharedByUserId string, sharedWithUserId string) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		query := `INSERT INTO CHECKLIST_SHARE(ID, CHECKLIST_ID, SHARED_BY_USER_ID, SHARED_WITH_USER_ID, PERMISSION_LEVEL, CREATED_AT)
				  VALUES (nextval('checklist_share_id_sequence'), @checklist_id, @shared_by, @shared_with, @permission_level, CURRENT_TIMESTAMP)
				  ON CONFLICT (CHECKLIST_ID, SHARED_WITH_USER_ID) DO NOTHING`

		result, err := tx.Exec(ctx, query, pgx.NamedArgs{
			"checklist_id":     checklistId,
			"shared_by":        sharedByUserId,
			"shared_with":      sharedWithUserId,
			"permission_level": "READ", // Default permission level
		})

		if err != nil {
			return false, err
		}

		return result.RowsAffected() > 0, nil
	}

	_, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Query:      queryFunc,
		Connection: repository.connection,
		TxOptions:  pgx.TxOptions{IsoLevel: pgx.Serializable},
	})

	if err != nil {
		return domain.Wrap(err, "Failed to create checklist share", 500)
	}

	return nil
}

func (repository *checklistRepository) DeleteChecklistShare(ctx context.Context, checklistId uint, userId string) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		query := `DELETE FROM CHECKLIST_SHARE
				  WHERE CHECKLIST_ID = @checklist_id
				    AND SHARED_WITH_USER_ID = @user_id`

		result, err := tx.Exec(ctx, query, pgx.NamedArgs{
			"checklist_id": checklistId,
			"user_id":      userId,
		})

		if err != nil {
			return false, err
		}

		return result.RowsAffected() == 1, nil
	}

	success, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Query:      queryFunc,
		Connection: repository.connection,
		TxOptions:  pgx.TxOptions{IsoLevel: pgx.Serializable},
	})

	if err != nil {
		return domain.Wrap(err, "Failed to delete checklist share", 500)
	}

	if !success {
		return domain.NewError("You do not have shared access to this checklist", 404)
	}

	return nil
}
