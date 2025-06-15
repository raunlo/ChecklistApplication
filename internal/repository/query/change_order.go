package query

import (
	"context"
	"errors"
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/mapper"
	"github.com/raunlo/pgx-with-automapper/pool"
)

// ChangeChecklistItemOrderQueryFunction Moves item to different order number
type ChangeChecklistItemOrderQueryFunction struct {
	newOrderNumber  uint
	checklistId     uint
	checklistItemId uint
	sortOrder       domain.SortOrder
}

func (c *ChangeChecklistItemOrderQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	type linkedItem struct {
		Id         uint  `primaryKey:"checklist_item_id"`
		NextItemId *uint `db:"next_item_id"`
	}
	// Finds checklist item by order number and returns item id and next item id
	findLinkedItemByOrderNumberFn := func(tx pool.TransactionWrapper) (linkedItem, error) {
		findLinkedItemByOrderNumberSQL := `WITH RECURSIVE CHECKLIST_ITEMS_CTE as (
			SELECT CHECKLIST_ID, CHECKLIST_ITEM_ID, CHECKLIST_ITEM_NAME, CHECKLIST_ITEM_COMPLETED, NEXT_ITEM_ID,
			1 as ORDER_NUMBER
			FROM CHECKLIST_ITEM WHERE CHECKLIST_ID = @checklist_id  AND NEXT_ITEM_ID  IS NULL
			
			UNION ALL
			
			SELECT CHECKLIST_ITEM.CHECKLIST_ID, CHECKLIST_ITEM.CHECKLIST_ITEM_ID, CHECKLIST_ITEM.CHECKLIST_ITEM_NAME, CHECKLIST_ITEM.CHECKLIST_ITEM_COMPLETED,
				CHECKLIST_ITEM.NEXT_ITEM_ID, ORDER_NUMBER + 1  as ORDER_NUMBER
			FROM CHECKLIST_ITEM, CHECKLIST_ITEMS_CTE
			WHERE CHECKLIST_ITEM.CHECKLIST_ID = @checklist_id  AND CHECKLIST_ITEMS_CTE.CHECKLIST_ITEM_ID =  CHECKLIST_ITEM.NEXT_ITEM_ID),

		    items_ordered as (
				SELECT CHECKLIST_ID, CHECKLIST_ITEM_ID, CHECKLIST_ITEM_NAME, CHECKLIST_ITEM_COMPLETED, NEXT_ITEM_ID, ORDER_NUMBER
				FROM CHECKLIST_ITEMS_CTE ORDER BY ORDER_NUMBER %s)

			SELECT 
				CHECKLIST_ID, CHECKLIST_ITEM_ID, CHECKLIST_ITEM_NAME, CHECKLIST_ITEM_COMPLETED, NEXT_ITEM_ID, ORDER_NUMBER
				from items_ordered where ORDER_NUMBER = (
																CASE WHEN @order_number = 1 THEN 1 ELSE
																	ABS(@order_number - 1) end)`

		findChecklistNextItemByOrderNumberArgs := pgx.NamedArgs{
			"checklist_id": c.checklistId,
			"order_number": c.newOrderNumber,
		}

		findLinkedItemByOrderNumberSQL = fmt.Sprintf(findLinkedItemByOrderNumberSQL, c.sortOrder.GetValue())

		var result linkedItem
		err := tx.QueryOne(context.Background(), findLinkedItemByOrderNumberSQL, &result, findChecklistNextItemByOrderNumberArgs)

		return result, err
	}

	updateChecklistItemPreviousItemOrderLinkFn := func(tx pool.TransactionWrapper, newNextChecklistItemId uint) (bool, error) {
		updateChecklistItemPreviousItemOrderLinkSql := `UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = @checklistItemId
				WHERE CHECKLIST_ID = @checklistId and
				CHECKLIST_ITEM_ID = (select CHECKLIST_ITEM_ID from CHECKLIST_ITEM where COALESCE(CHECKLIST_ITEM_ID, 0) = COALESCE(@newNextChecklistItemId, 0) and checklist_id = @checklistId)`

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

	updateChecklistItemOrderLinkFn := func(tx pool.TransactionWrapper, newNextItemId *uint) (bool, error) {
		updateChecklistItemOrderLinkSql := `UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = @newNextItemId
                WHERE CHECKLIST_ID = @checklistId  and CHECKLIST_ITEM_ID =  @checklistItemId`

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
		checklistItemIdAndNextITemId, err := findLinkedItemByOrderNumberFn(tx)
		if err != nil && errors.Is(err, mapper.ErrNoRows) {
			return false, errors.New("no checklist item found with order number")
		} else if err != nil {
			return false, err
		}

		if checklistItemIdAndNextITemId.Id == c.checklistItemId {
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
		ok, err = updateChecklistItemPreviousItemOrderLinkFn(tx, checklistItemIdAndNextITemId.Id)
		if err != nil || !ok {
			return false, err
		}
		ok, err = updateChecklistItemOrderLinkFn(tx, checklistItemIdAndNextITemId.NextItemId)
		return ok, err
	}
}
