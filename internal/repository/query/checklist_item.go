package query

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/repository/dbo"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/raunlo/pgx-with-automapper/pool"
)

// PersistChecklistItemQueryFunction Persist checklistItem struct using gap-based positioning
type PersistChecklistItemQueryFunction struct {
	checklistId   uint
	checklistItem domain.ChecklistItem
}

func (p *PersistChecklistItemQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (domain.ChecklistItem, error) {
	return func(tx pool.TransactionWrapper) (domain.ChecklistItem, error) {
		// Get the maximum position for uncompleted items (new items go at the end)
		var maxPosition float64
		err := tx.QueryRow(context.Background(),
			`SELECT COALESCE(MAX(POSITION), 0) FROM CHECKLIST_ITEM
			 WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_COMPLETED = FALSE`,
			pgx.NamedArgs{"checklistId": p.checklistId}).Scan(&maxPosition)
		if err != nil {
			return domain.ChecklistItem{}, err
		}

		newPosition := maxPosition + domain.DefaultGapSize

		// Insert new item at the end
		insertSql := `INSERT INTO CHECKLIST_ITEM(CHECKLIST_ITEM_ID, CHECKLIST_ID, CHECKLIST_ITEM_NAME, CHECKLIST_ITEM_COMPLETED, POSITION, UPDATED_AT)
					  VALUES(nextval('checklist_item_id_sequence'), @checklistId, @checklistItemName, @checklistItemCompleted, @position, CURRENT_TIMESTAMP)
					  RETURNING CHECKLIST_ITEM_ID`

		err = tx.QueryRow(context.Background(), insertSql, pgx.NamedArgs{
			"checklistId":            p.checklistId,
			"checklistItemName":      p.checklistItem.Name,
			"checklistItemCompleted": p.checklistItem.Completed,
			"position":               newPosition,
		}).Scan(&p.checklistItem.Id)

		if err != nil {
			return domain.ChecklistItem{}, err
		}

		p.checklistItem.Position = newPosition
		return p.checklistItem, nil
	}
}

// GetAllChecklistItemsQueryFunction Get all checklist queries struct
type GetAllChecklistItemsQueryFunction struct {
	checklistId uint
	completed   *bool
	sortOrder   domain.SortOrder
}

func (p *GetAllChecklistItemsQueryFunction) GetQueryFunction(ctx context.Context) func(connection pool.Conn) ([]dbo.ChecklistItemDbo, error) {
	return func(connection pool.Conn) ([]dbo.ChecklistItemDbo, error) {
		query := `
			SELECT
				ci.CHECKLIST_ITEM_ID,
				ci.CHECKLIST_ITEM_NAME,
				ci.CHECKLIST_ITEM_COMPLETED,
				ci.POSITION,
				ROW_NUMBER() OVER (
					PARTITION BY ci.CHECKLIST_ID
					ORDER BY ci.CHECKLIST_ITEM_COMPLETED ASC, ci.POSITION ASC
				) AS ORDER_NUMBER,
				ROWS.CHECKLIST_ITEM_ROW_ID,
				ROWS.CHECKLIST_ITEM_ROW_NAME,
				ROWS.CHECKLIST_ITEM_ROW_COMPLETED
			FROM CHECKLIST_ITEM ci
			LEFT JOIN CHECKLIST_ITEM_ROW AS ROWS ON ROWS.CHECKLIST_ITEM_ID = ci.CHECKLIST_ITEM_ID
			WHERE (CAST(@checklist_item_completed as Boolean) IS NULL OR ci.CHECKLIST_ITEM_COMPLETED = @checklist_item_completed)
			  AND ci.CHECKLIST_ID = @checklist_id
			ORDER BY ci.CHECKLIST_ITEM_COMPLETED ASC, ci.POSITION ASC, ROWS.CHECKLIST_ITEM_ROW_COMPLETED ASC`

		var result []dbo.ChecklistItemDbo
		err := connection.QueryList(context.Background(), query, &result, pgx.NamedArgs{
			"checklist_id":             p.checklistId,
			"checklist_item_completed": p.completed,
		})

		return result, err
	}
}

// DeleteChecklistItemQueryFunction Delete checklist item by id query struct
type DeleteChecklistItemQueryFunction struct {
	checklistId     uint
	checklistItemId uint
}

func (d *DeleteChecklistItemQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		// Lock the row to prevent concurrent deletes
		lockSQL := `SELECT CHECKLIST_ITEM_ID FROM CHECKLIST_ITEM
					WHERE CHECKLIST_ID = @checklist_id AND CHECKLIST_ITEM_ID = @checklist_item_id
					FOR UPDATE`

		var itemId uint
		err := tx.QueryRow(context.Background(), lockSQL, pgx.NamedArgs{
			"checklist_item_id": d.checklistItemId,
			"checklist_id":      d.checklistId,
		}).Scan(&itemId)
		if err != nil {
			if err == pgx.ErrNoRows {
				return false, nil
			}
			return false, err
		}

		// Delete the item
		removeChecklistItemSQL := `DELETE FROM CHECKLIST_ITEM
					WHERE CHECKLIST_ID = @checklist_id AND CHECKLIST_ITEM_ID = @checklist_item_id`

		result, err := tx.Exec(context.Background(), removeChecklistItemSQL, pgx.NamedArgs{
			"checklist_item_id": d.checklistItemId,
			"checklist_id":      d.checklistId,
		})

		if err != nil {
			return false, err
		}

		if result.RowsAffected() > 1 {
			return false, errors.New("removeChecklistItem affected more than one row")
		}
		return result.RowsAffected() == 1, nil
	}
}

// FindChecklistItemById Find checklist by id query struct
type FindChecklistItemById struct {
	checklistId     uint
	checklistItemId uint
}

func (f *FindChecklistItemById) GetQueryFunction(ctx context.Context) func(connection pool.Conn) (*dbo.ChecklistItemDbo, error) {
	return func(connection pool.Conn) (*dbo.ChecklistItemDbo, error) {
		sql := `SELECT
					ci.CHECKLIST_ITEM_ID,
					ci.CHECKLIST_ITEM_NAME,
					ci.CHECKLIST_ITEM_COMPLETED,
					ci.POSITION,
					CIR.CHECKLIST_ITEM_ROW_NAME,
					CIR.CHECKLIST_ITEM_ROW_COMPLETED,
					CIR.CHECKLIST_ITEM_ROW_ID
				FROM CHECKLIST_ITEM ci
				LEFT JOIN CHECKLIST_ITEM_ROW CIR ON ci.CHECKLIST_ITEM_ID = CIR.CHECKLIST_ITEM_ID
				WHERE ci.CHECKLIST_ID = @checklistId AND ci.CHECKLIST_ITEM_ID = @checklistItemId`

		args := pgx.NamedArgs{
			"checklistId":     f.checklistId,
			"checklistItemId": f.checklistItemId,
		}
		var dboItem dbo.ChecklistItemDbo
		err := connection.QueryOne(context.Background(), sql, &dboItem, args)

		return &dboItem, err
	}
}

// UpdateChecklistItemFunction update checklist item query struct
type UpdateChecklistItemFunction struct {
	checklistId   uint
	checklistItem domain.ChecklistItem
}

func (u *UpdateChecklistItemFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		// Lock the row to prevent concurrent updates
		lockSQL := `SELECT CHECKLIST_ITEM_ID FROM CHECKLIST_ITEM
					WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @checklistItemId
					FOR UPDATE`

		var itemId uint
		err := tx.QueryRow(context.Background(), lockSQL, pgx.NamedArgs{
			"checklistId":     u.checklistId,
			"checklistItemId": u.checklistItem.Id,
		}).Scan(&itemId)

		if err != nil {
			return false, err
		}

		// Perform the update
		sql := `UPDATE CHECKLIST_ITEM
				SET CHECKLIST_ITEM_NAME = @checklistItemName, CHECKLIST_ITEM_COMPLETED = @checklistItemCompleted, UPDATED_AT = CURRENT_TIMESTAMP
				WHERE CHECKLIST_ID = @checklistId and CHECKLIST_ITEM_ID = @checklistItemId`

		args := pgx.NamedArgs{
			"checklistItemName":      u.checklistItem.Name,
			"checklistItemCompleted": u.checklistItem.Completed,
			"checklistId":            u.checklistId,
			"checklistItemId":        u.checklistItem.Id,
		}
		res, err := tx.Exec(context.Background(), sql, args)

		return res.RowsAffected() == 1, err
	}
}
