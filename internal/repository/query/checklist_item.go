package query

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/repository/dbo"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type RemoveOrderLinkQueryFunction struct {
	checklistId     uint
	checklistItemId uint
}

func (r *RemoveOrderLinkQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		removeChecklistItemFromOrderLinkSQL := `UPDATE CHECKLIST_ITEM
		SET NEXT_ITEM_ID = (SELECT NEXT_ITEM_ID FROM CHECKLIST_ITEM WHERE CHECKLIST_ID =  @checklistId and CHECKLIST_ITEM_ID = @checklistItemId)
		WHERE CHECKLIST_ID = @checklistId AND NEXT_ITEM_ID = @checklistItemId `

		//UPDATE CHECKLIST_ITEM
		//SET NEXT_ITEM_ID = (SELECT NEXT_ITEM_ID FROM CHECKLIST_ITEM WHERE CHECKLIST_ID =  2 and CHECKLIST_ITEM_ID = 8)
		//WHERE CHECKLIST_ID = 2 AND NEXT_ITEM_ID = 8 `

		removeChecklistItemFromOrderLinkArgs := pgx.NamedArgs{
			"checklistId":     r.checklistId,
			"checklistItemId": r.checklistItemId,
		}

		tag, err := tx.Exec(context.Background(), removeChecklistItemFromOrderLinkSQL, removeChecklistItemFromOrderLinkArgs)
		if tag.RowsAffected() > 1 {
			return false, errors.New("removeTaskFromOrderLink was not updating only one row")
		}
		return true, err
	}
}

// PersistChecklistItemQueryFunction Persist checklistItem struct
type PersistChecklistItemQueryFunction struct {
	checklistId   uint
	checklistItem domain.ChecklistItem
}

func (p *PersistChecklistItemQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (domain.ChecklistItem, error) {
	updateOrderLinkFn := func(tx pool.TransactionWrapper, savedChecklistItemId uint) error {
		updateOrderLinkSQL := `UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = @new_checklist_item_id 
                               WHERE CHECKLIST_ID = @checklist_id AND NEXT_ITEM_ID IS NULL AND CHECKLIST_ITEM_ID <> @new_checklist_item_id`
		updateOrderLinkArgs := pgx.NamedArgs{
			"new_checklist_item_id": savedChecklistItemId,
			"checklist_id":          p.checklistId,
		}

		tags, err := tx.Exec(context.Background(), updateOrderLinkSQL, updateOrderLinkArgs)
		if tags.RowsAffected() > 1 {
			return errors.New("UpdateOrderLink function updated more than one row")
		}
		return err
	}

	insertChecklistItemFn := func(tx pool.TransactionWrapper) (domain.ChecklistItem, error) {
		insertSql := `INSERT INTO CHECKLIST_ITEM(CHECKLIST_ITEM_ID, CHECKLIST_ID, CHECKLIST_ITEM_NAME, CHECKLIST_ITEM_COMPLETED, NEXT_ITEM_ID) 
				VALUES(nextval('checklist_item_id_sequence'), @checklistId, @checklistItemName, @checklistItemCompleted, @checklistItemNextId)
				RETURNING CHECKLIST_ITEM_ID `
		insertSQLArgs := pgx.NamedArgs{
			"checklistId":            p.checklistId,
			"checklistItemName":      p.checklistItem.Name,
			"checklistItemNextId":    nil,
			"checklistItemCompleted": p.checklistItem.Completed,
		}

		row := tx.QueryRow(context.Background(), insertSql, insertSQLArgs)

		err := row.Scan(&p.checklistItem.Id)
		return p.checklistItem, err
	}

	return func(tx pool.TransactionWrapper) (domain.ChecklistItem, error) {
		checklistItem, err := insertChecklistItemFn(tx)
		if err != nil {
			return domain.ChecklistItem{}, err
		} else {
			err = updateOrderLinkFn(tx, checklistItem.Id)
		}

		return checklistItem, err
	}
}

// GetAllChecklistItemsQueryFunction Get all checklist queries struct
type GetAllChecklistItemsQueryFunction struct {
	checklistId uint
	completed   *bool
	sortOrder   domain.SortOrder
}

func (p *GetAllChecklistItemsQueryFunction) GetQueryFunction() func(connection pool.Conn) ([]dbo.ChecklistItemDbo, error) {
	return func(connection pool.Conn) ([]dbo.ChecklistItemDbo, error) {
		query := `WITH RECURSIVE CHECKLIST_ITEMS_CTE as (
			SELECT CHECKLIST_ID, CHECKLIST_ITEM_ID, CHECKLIST_ITEM_NAME, CHECKLIST_ITEM_COMPLETED, NEXT_ITEM_ID,
			1 as ORDER_NUMBER
			FROM CHECKLIST_ITEM WHERE CHECKLIST_ID = @checklist_id  AND NEXT_ITEM_ID IS NULL AND
			(CHECKLIST_ITEM_COMPLETED IS NOT NULL AND CHECKLIST_ITEM_COMPLETED = @checklist_item_completed)
			
			UNION ALL
			
			SELECT CHECKLIST_ITEM.CHECKLIST_ID, CHECKLIST_ITEM.CHECKLIST_ITEM_ID, CHECKLIST_ITEM.CHECKLIST_ITEM_NAME, CHECKLIST_ITEM.CHECKLIST_ITEM_COMPLETED, CHECKLIST_ITEM.NEXT_ITEM_ID, ORDER_NUMBER + 1  as ORDER_NUMBER
			FROM CHECKLIST_ITEM, CHECKLIST_ITEMS_CTE
			WHERE CHECKLIST_ITEM.CHECKLIST_ID = @checklist_id  AND CHECKLIST_ITEMS_CTE.CHECKLIST_ITEM_ID =  CHECKLIST_ITEM.NEXT_ITEM_ID
			AND (CHECKLIST_ITEM.CHECKLIST_ITEM_COMPLETED IS NOT NULL AND CHECKLIST_ITEM.CHECKLIST_ITEM_COMPLETED = @checklist_item_completed))
					
			SELECT CHECKLIST_ITEMS_CTE.CHECKLIST_ITEM_ID, CHECKLIST_ITEMS_CTE.CHECKLIST_ITEM_NAME, CHECKLIST_ITEMS_CTE.CHECKLIST_ITEM_COMPLETED, ORDER_NUMBER,
			ROWS.CHECKLIST_ITEM_ROW_ID, ROWS.CHECKLIST_ITEM_ROW_NAME, ROWS.CHECKLIST_ITEM_ROW_COMPLETED  FROM CHECKLIST_ITEMS_CTE
			LEFT JOIN CHECKLIST_ITEM_ROW AS ROWS   on  ROWS.CHECKLIST_ITEM_ID =  CHECKLIST_ITEMS_CTE.CHECKLIST_ITEM_ID ORDER BY ORDER_NUMBER `

		query += p.sortOrder.GetValue()
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
	removeChecklistItemFromOrderLinkFn := func(tx pool.TransactionWrapper) error {
		queryFunction := RemoveOrderLinkQueryFunction{
			checklistItemId: d.checklistItemId,
			checklistId:     d.checklistId,
		}
		_, err := queryFunction.GetTransactionalQueryFunction()(tx)
		return err
	}

	removeChecklistItemFn := func(tx pgx.Tx) (bool, error) {
		removeChecklistItemSQL := `DELETE FROM CHECKLIST_ITEM
       				 WHERE CHECKLIST_ID = @checklist_id AND CHECKLIST_ITEM_ID = @checklist_item_id`
		removeChecklistItemParams := pgx.NamedArgs{
			"CHECKLIST_ITEM_ID": d.checklistItemId,
			"CHECKLIST_ID":      d.checklistId,
		}

		result, err := tx.Exec(context.Background(), removeChecklistItemSQL, removeChecklistItemParams)

		if result.RowsAffected() > 1 {
			return false, errors.New("removeChecklistItem did not affect more than one row")
		}
		return result.RowsAffected() == 1, err

	}
	return func(tx pool.TransactionWrapper) (bool, error) {
		var result bool
		err := removeChecklistItemFromOrderLinkFn(tx)
		if err == nil {
			result, err = removeChecklistItemFn(tx)
		}
		return result, err
	}
}

// FindChecklistItemById Find checklist by id query struct
type FindChecklistItemById struct {
	checklistId     uint
	checklistItemId uint
}

func (f *FindChecklistItemById) GetQueryFunction() func(connection pool.Conn) (*dbo.ChecklistItemDbo, error) {
	return func(connection pool.Conn) (*dbo.ChecklistItemDbo, error) {
		sql := `SELECT CHECKLIST_ITEM.CHECKLIST_ITEM_ID, CHECKLIST_ITEM.CHECKLIST_ITEM_NAME, CHECKLIST_ITEM.CHECKLIST_ITEM_COMPLETED,
       			NEXT_ITEM_ID, CIR.CHECKLIST_ITEM_ROW_NAME, cir.CHECKLIST_ITEM_ROW_COMPLETED
       			FROM CHECKLIST_ITEM
                LEFT JOIN CHECKLIST_ITEM_ROW cir on CHECKLIST_ITEM.CHECKLIST_ITEM_ID = cir.CHECKLIST_ITEM_ID
         		WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM.CHECKLIST_ITEM_ID = @checklistItemId`

		args := pgx.NamedArgs{
			"checklistId":     f.checklistId,
			"checklistItemId": f.checklistItemId,
		}
		var dbo dbo.ChecklistItemDbo
		err := connection.QueryOne(context.Background(), sql, &dbo, args)

		return &dbo, err
	}
}

// UpdateChecklistItemFunction update checklist item query struct
type UpdateChecklistItemFunction struct {
	checklistId   uint
	checklistItem domain.ChecklistItem
}

func (u *UpdateChecklistItemFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		sql := `UPDATE CHECKLIST_ITEM
				SET name = @checklistItemName AND COMPLETED = @checklistItemCompleted
				WHERE CHECKLIST_ID = @checklistId and ID = @checklistItemId
		
		`
		args := pgx.NamedArgs{
			"checklistId":            u.checklistId,
			"checklistItemId":        u.checklistItem.Id,
			"checklistItemName":      u.checklistItem.Name,
			"checklistItemCompleted": u.checklistItem.Completed,
		}
		res, err := tx.Exec(context.Background(), sql, args)

		return res.RowsAffected() == 1, err
	}
}
