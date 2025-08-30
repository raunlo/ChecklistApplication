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
		// fetch current links for the item
		var prevItemId *uint
		var nextItemId *uint
		row := tx.QueryRow(context.Background(), `SELECT PREV_ITEM_ID, NEXT_ITEM_ID FROM CHECKLIST_ITEM
                                WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @checklistItemId FOR UPDATE`, pgx.NamedArgs{
			"checklistId":     r.checklistId,
			"checklistItemId": r.checklistItemId,
		})
		if err := row.Scan(&prevItemId, &nextItemId); err != nil {
			return false, err
		}

		// detach moving item from neighbours
		_, err := tx.Exec(context.Background(), `UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = NULL, PREV_ITEM_ID = NULL
                                WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @checklistItemId`, pgx.NamedArgs{
			"checklistId":     r.checklistId,
			"checklistItemId": r.checklistItemId,
		})
		if err != nil {
			return false, err
		}

		if prevItemId != nil {
			tag, err := tx.Exec(context.Background(), `UPDATE CHECKLIST_ITEM
                                        SET NEXT_ITEM_ID = @nextItemId
                                        WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @prevItemId`, pgx.NamedArgs{
				"checklistId": r.checklistId,
				"nextItemId":  nextItemId,
				"prevItemId":  prevItemId,
			})
			if err != nil {
				return false, err
			} else if tag.RowsAffected() > 1 {
				return false, errors.New("removeTaskFromOrderLink was not updating only one row")
			}
		}

		if nextItemId != nil {
			tag, err := tx.Exec(context.Background(), `UPDATE CHECKLIST_ITEM
                                        SET PREV_ITEM_ID = @prevItemId
                                        WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @nextItemId`, pgx.NamedArgs{
				"checklistId": r.checklistId,
				"prevItemId":  prevItemId,
				"nextItemId":  nextItemId,
			})
			if err != nil {
				return false, err
			} else if tag.RowsAffected() > 1 {
				return false, errors.New("removeTaskFromOrderLink was not updating only one row")
			}
		}

		return true, nil

	}
}

// PersistChecklistItemQueryFunction Persist checklistItem struct
type PersistChecklistItemQueryFunction struct {
	checklistId   uint
	checklistItem domain.ChecklistItem
}

func (p *PersistChecklistItemQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (domain.ChecklistItem, error) {
	queryPhantomElementId := func(tx pool.TransactionWrapper) (uint, *uint, error) {
		sql := `
				SELECT CHECKLIST_ITEM_ID, NEXT_ITEM_ID FROM CHECKLIST_ITEM
				WHERE CHECKLIST_ID = @checklistId and PREV_ITEM_ID IS NULL AND IS_PHANTOM = TRUE
				FOR UPDATE
			`
		arguments := pgx.NamedArgs{
			"checklistId": p.checklistId,
		}
		row := tx.QueryRow(context.Background(), sql, arguments)

		var phantomElementId uint
		var phantomElementNextItemId *uint
		err := row.Scan(&phantomElementId, &phantomElementNextItemId)
		return phantomElementId, phantomElementNextItemId, err
	}

	setPhantomNextItemToNewlyCreatedItem := func(tx pool.TransactionWrapper, newlyCreatedItemId uint, phantomElementId uint) error {
		updatePrevItemOrderLinkSQL := `UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = @newlyCreatedItemId 
									WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @phantomElementId`

		_, err := tx.Exec(context.Background(), updatePrevItemOrderLinkSQL, pgx.NamedArgs{
			"newlyCreatedItemId": newlyCreatedItemId,
			"checklistId":        p.checklistId,
			"phantomElementId":   phantomElementId,
		})

		return err
	}

	insertChecklistItemFn := func(tx pool.TransactionWrapper, phantomElementId uint, phantomElemntNextItemId *uint) (domain.ChecklistItem, error) {
		insertSql := `INSERT INTO CHECKLIST_ITEM(CHECKLIST_ITEM_ID, CHECKLIST_ID, CHECKLIST_ITEM_NAME, CHECKLIST_ITEM_COMPLETED, NEXT_ITEM_ID, PREV_ITEM_ID) 
				VALUES(nextval('checklist_item_id_sequence'), @checklistId, @checklistItemName, @checklistItemCompleted, @checklistItemNextId, @checklistItemPrevId)
				RETURNING CHECKLIST_ITEM_ID`
		insertSQLArgs := pgx.NamedArgs{
			"checklistId":            p.checklistId,
			"checklistItemName":      p.checklistItem.Name,
			"checklistItemNextId":    phantomElemntNextItemId,
			"checklistItemPrevId":    phantomElementId,
			"checklistItemCompleted": p.checklistItem.Completed,
		}

		row := tx.QueryRow(context.Background(), insertSql, insertSQLArgs)

		err := row.Scan(&p.checklistItem.Id)
		return p.checklistItem, err
	}

	return func(tx pool.TransactionWrapper) (domain.ChecklistItem, error) {
		phantomItemId, phantomElementNextItemId, err := queryPhantomElementId(tx)

		if err != nil {
			return domain.ChecklistItem{}, err
		}
		checklistItem, err := insertChecklistItemFn(tx, phantomItemId, phantomElementNextItemId)
		if err != nil {
			return domain.ChecklistItem{}, err
		}
		err = setPhantomNextItemToNewlyCreatedItem(tx, checklistItem.Id, phantomItemId)

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
		query := `				
			SELECT CHECKLIST_ITEMS_ORDERED_VIEW.CHECKLIST_ITEM_ID, CHECKLIST_ITEM_NAME, CHECKLIST_ITEM_COMPLETED, ORDER_NUMBER,
			ROWS.CHECKLIST_ITEM_ROW_ID, ROWS.CHECKLIST_ITEM_ROW_NAME, ROWS.CHECKLIST_ITEM_ROW_COMPLETED  FROM CHECKLIST_ITEMS_ORDERED_VIEW
			LEFT JOIN CHECKLIST_ITEM_ROW AS ROWS   on  ROWS.CHECKLIST_ITEM_ID =  CHECKLIST_ITEMS_ORDERED_VIEW.CHECKLIST_ITEM_ID 
			WHERE (CAST(@checklist_item_completed as Boolean) IS NULL OR CHECKLIST_ITEM_COMPLETED = @checklist_item_completed) AND CHECKLIST_ID = @checklist_id
			ORDER BY CHECKLIST_ITEM_COMPLETED ASC, ORDER_NUMBER ASC, CHECKLIST_ITEM_ROW_COMPLETED ASC`

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
			"checklist_item_id": d.checklistItemId,
			"checklist_id":      d.checklistId,
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
		sql := `SELECT CHECKLIST_ITEMS_ORDERED_VIEW.CHECKLIST_ITEM_ID, CHECKLIST_ITEMS_ORDERED_VIEW.CHECKLIST_ITEM_NAME, CHECKLIST_ITEMS_ORDERED_VIEW.CHECKLIST_ITEM_COMPLETED,
       			CIR.CHECKLIST_ITEM_ROW_NAME, cir.CHECKLIST_ITEM_ROW_COMPLETED, cir.CHECKLIST_ITEM_ROW_ID
       			FROM CHECKLIST_ITEMS_ORDERED_VIEW
                LEFT JOIN CHECKLIST_ITEM_ROW cir on CHECKLIST_ITEMS_ORDERED_VIEW.CHECKLIST_ITEM_ID = cir.CHECKLIST_ITEM_ID
         		WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEMS_ORDERED_VIEW.CHECKLIST_ITEM_ID = @checklistItemId`

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
				SET CHECKLIST_ITEM_NAME = @checklistItemName, CHECKLIST_ITEM_COMPLETED = @checklistItemCompleted
				WHERE CHECKLIST_ID = @checklistId and CHECKLIST_ITEM_ID = @checklistItemId
		
		`
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
