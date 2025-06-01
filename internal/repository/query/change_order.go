package query

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/pool"
)

// ChangeChecklistItemOrderQueryFunction Moves item to different order number
type ChangeChecklistItemOrderQueryFunction struct {
	newOrderNumber  uint
	checklistId     uint
	checklistItemId uint
	SortType        domain.SortOrder
}

func (c *ChangeChecklistItemOrderQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	findChecklistNextItemByOrderNumberFn := func(tx pool.TransactionWrapper) (uint, error) {
		findChecklistNextItemByOrderNumberQuery := `WITH RECURSIVE CHECKLIST_ITEMS_CTE as (
			SELECT CHECKLIST_ID, ID, NAME, COMPLETED, NEXT_ITEM_ID,
			1 as ORDER_NUMBER
			FROM CHECKLIST_ITEM WHERE CHECKLIST_ID = @checklist_id  AND NEXT_ITEM_ID  IS NULL
			
			UNION ALL
			
			SELECT CHECKLIST_ITEM.CHECKLIST_ID, CHECKLIST_ITEM.ID, CHECKLIST_ITEM.NAME, CHECKLIST_ITEM.COMPLETED, CHECKLIST_ITEM.NEXT_ITEM_ID, ORDER_NUMBER + 1  as ORDER_NUMBER
			FROM CHECKLIST_ITEM, CHECKLIST_ITEMS_CTE
			WHERE CHECKLIST_ITEM.CHECKLIST_ID = @checklist_id  AND CHECKLIST_ITEMS_CTE.ID =  CHECKLIST_ITEM.NEXT_ITEM_ID),
		    items_ordered as (select CHECKLIST_ID, ID, NAME, COMPLETED, NEXT_ITEM_ID, ORDER_NUMBER from CHECKLIST_ITEMS_CTE ORDER BY ORDER_NUMBER %s)
			SELECT 
				CASE when @order_number = 1 THEN ID ELSE NEXT_ITEM_ID END
				from items_ordered where ORDER_NUMBER = (
																CASE WHEN @order_number = 1 THEN 1 ELSE
																	abs(@order_number - 1) end)`

		findChecklistNextItemByOrderNumberArgs := pgx.NamedArgs{
			"checklist_id": c.checklistId,
			"order_number": c.newOrderNumber,
		}

		findChecklistNextItemByOrderNumberQuery = fmt.Sprintf(findChecklistNextItemByOrderNumberQuery, c.SortType.GetValue())

		row := tx.QueryRow(context.Background(), findChecklistNextItemByOrderNumberQuery, findChecklistNextItemByOrderNumberArgs)
		var checklistItemId uint

		err := row.Scan(&checklistItemId)

		return checklistItemId, err

	}

	updateChecklistItemPreviousItemOrderLinkFn := func(tx pool.TransactionWrapper, newNextChecklistItemId uint) (bool, error) {
		updateChecklistItemPreviousItemOrderLinkSql := `UPDATE checklist_item SET NEXT_ITEM_ID = @checklistItemId
				WHERE checklist_id = @checklistId and
				ID = (select ID from checklist_item where COALESCE(NEXT_ITEM_ID, 0) = COALESCE(@newNextChecklistItemId, 0) and checklist_id = @checklistId)`

		updateChecklistItemPreviousItemOrderLinkArgs := pgx.NamedArgs{
			"checklistId":            c.checklistId,
			"checklistItemId":        c.checklistItemId,
			"newNextChecklistItemId": newNextChecklistItemId,
		}

		tag, err := tx.Exec(context.Background(), updateChecklistItemPreviousItemOrderLinkSql,
			updateChecklistItemPreviousItemOrderLinkArgs)
		if err != nil {
			return false, err
		} else if tag.RowsAffected() > 1 {
			return false, errors.New("updateChecklistItemPreviousItemOrderLinkFn affected more than one row")
		}

		return tag.RowsAffected() == 1, err
	}

	updateChecklistItemOrderLinkFn := func(tx pool.TransactionWrapper, newNextItemId uint) (bool, error) {
		updateChecklistItemOrderLinkSql := `UPDATE checklist_item SET NEXT_ITEM_ID = @newNextItemId
                WHERE checklist_id = @checklistId  and ID =  @checklistItemId`

		updateChecklistItemOrderLinkArgs := pgx.NamedArgs{
			"newNextItemId":   newNextItemId,
			"checklistId":     c.checklistId,
			"checklistItemId": c.checklistItemId,
		}

		tag, err := tx.Exec(context.Background(), updateChecklistItemOrderLinkSql, updateChecklistItemOrderLinkArgs)

		if tag.RowsAffected() > 1 {
			return false, errors.New("updateChecklistItemOrderLinkFn Affected more than one row")
		}
		return tag.RowsAffected() == 1, err
	}

	return func(tx pool.TransactionWrapper) (bool, error) {
		checklistItemId, err := findChecklistNextItemByOrderNumberFn(tx)
		if err != nil && errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("no checklist item found with order number")
		} else if err != nil {
			return false, err
		}

		if checklistItemId == c.checklistItemId {
			return true, nil
		}

		removeChecklistItemOrderLink := RemoveOrderLinkQueryFunction{
			checklistId:     c.checklistId,
			checklistItemId: c.checklistItemId,
		}
		ok, err := removeChecklistItemOrderLink.GetTransactionalQueryFunction()(tx)
		if err != nil || !ok {
			return false, err
		}
		ok, err = updateChecklistItemPreviousItemOrderLinkFn(tx, checklistItemId)
		if err != nil || !ok {
			return false, err
		}
		ok, err = updateChecklistItemOrderLinkFn(tx, checklistItemId)
		return ok, err
	}
}
