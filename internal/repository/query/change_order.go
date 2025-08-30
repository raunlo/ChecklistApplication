package query

import (
	"context"
	"errors"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/jackc/pgx/v5"
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
	return func(tx pool.TransactionWrapper) (bool, error) {
		removeChecklistItemOrderLink := RemoveOrderLinkQueryFunction{
			checklistId:     c.checklistId,
			checklistItemId: c.checklistItemId,
		}
		ok, err := removeChecklistItemOrderLink.GetTransactionalQueryFunction()(tx)
		if err != nil || !ok {
			return false, err
		}

		itemId, nextItemId, err := c.findDesiredItemWithNextAndPreviousLinksByOrderNumber(tx)
		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			return false, errors.New("no checklist item found with order number")
		} else if err != nil {
			return false, err
		}

		return c.linkChecklistItemAtNewPosition(tx, itemId, nextItemId)
	}
}

// Finds item by order number and returns item id and and next item
func (c *ChangeChecklistItemOrderQueryFunction) findDesiredItemWithNextAndPreviousLinksByOrderNumber(tx pool.TransactionWrapper) (*uint, *uint, error) {
	findLinkedItemByOrderNumberSQL := `WITH RECURSIVE CHECKLIST_ITEMS_CTE as (
			SELECT CHECKLIST_ITEM_ID, NEXT_ITEM_ID, PREV_ITEM_ID,1 as ORDER_NUMBER
			FROM CHECKLIST_ITEM WHERE CHECKLIST_ID = @checklistId  AND NEXT_ITEM_ID IS NULL AND CHECKLIST_ITEM_ID <> @itemToMoveId
			
			UNION ALL
			
			SELECT CHECKLIST_ITEM.CHECKLIST_ITEM_ID, CHECKLIST_ITEM.NEXT_ITEM_ID, CHECKLIST_ITEM.PREV_ITEM_ID, ORDER_NUMBER + 1  as ORDER_NUMBER
			FROM CHECKLIST_ITEM, CHECKLIST_ITEMS_CTE
			WHERE CHECKLIST_ITEM.CHECKLIST_ID = @checklistId  AND CHECKLIST_ITEMS_CTE.CHECKLIST_ITEM_ID =  CHECKLIST_ITEM.NEXT_ITEM_ID)

			SELECT CHECKLIST_ITEM_ID, PREV_ITEM_ID, NEXT_ITEM_ID from CHECKLIST_ITEMS_CTE where ORDER_NUMBER = @orderNumber`

	findChecklistNextItemByOrderNumberArgs := pgx.NamedArgs{
		"checklistId":  c.checklistId,
		"orderNumber":  c.newOrderNumber,
		"itemToMoveId": c.checklistItemId,
	}

	var itemId uint
	var prevItemId *uint
	var nextItemId *uint
	row := tx.QueryRow(context.Background(), findLinkedItemByOrderNumberSQL, findChecklistNextItemByOrderNumberArgs)

	err := row.Scan(&itemId, &prevItemId, &nextItemId)

	return &itemId, nextItemId, err
}

func (c *ChangeChecklistItemOrderQueryFunction) linkChecklistItemAtNewPosition(tx pool.TransactionWrapper, newPrevItemId *uint, newNextItemId *uint) (bool, error) {
	if newPrevItemId != nil {
		tag, err := tx.Exec(context.Background(), `UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = @itemToMoveId
                                WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @newPrevItemId`, pgx.NamedArgs{
			"checklistId":   c.checklistId,
			"itemToMoveId":  c.checklistItemId,
			"newPrevItemId": newPrevItemId,
		})
		if err != nil {
			return false, err
		} else if tag.RowsAffected() > 1 {
			return false, errors.New("updateChecklistItemPreviousItemOrderLinkFn affected more than one row")
		}
	}

	if newNextItemId != nil {
		tag, err := tx.Exec(context.Background(), `UPDATE CHECKLIST_ITEM SET PREV_ITEM_ID = @itemToMoveId
                                WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @newNextItemId`, pgx.NamedArgs{
			"checklistId":   c.checklistId,
			"itemToMoveId":  c.checklistItemId,
			"newNextItemId": newNextItemId,
		})
		if err != nil {
			return false, err
		} else if tag.RowsAffected() > 1 {
			return false, errors.New("updateChecklistItemPreviousItemOrderLinkFn affected more than one row")
		}
	}

	tag, err := tx.Exec(context.Background(), `UPDATE CHECKLIST_ITEM SET NEXT_ITEM_ID = @newNextItemId
                        WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @itemToMoveId`, pgx.NamedArgs{
		"checklistId":   c.checklistId,
		"itemToMoveId":  c.checklistItemId,
		"newNextItemId": newNextItemId,
	})
	if err != nil {
		return false, err
	} else if tag.RowsAffected() > 1 {
		return false, errors.New("updateChecklistItemPreviousItemOrderLinkFn affected more than one row")
	}

	tag, err = tx.Exec(context.Background(), `UPDATE CHECKLIST_ITEM SET PREV_ITEM_ID = @newPrevItemId
                        WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @itemToMoveId`, pgx.NamedArgs{
		"checklistId":   c.checklistId,
		"itemToMoveId":  c.checklistItemId,
		"newPrevItemId": newPrevItemId,
	})
	if err != nil {
		return false, err
	} else if tag.RowsAffected() > 1 {
		return false, errors.New("updateChecklistItemPreviousItemOrderLinkFn affected more than one row")
	}

	return true, nil
}
