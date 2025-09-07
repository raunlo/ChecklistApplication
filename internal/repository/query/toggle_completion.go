package query

import (
	"context"
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/pool"
)

// toggleCompletionQueryFunction is a struct for toggling item completion status
type toggleCompletionQueryFunction struct {
	checklistId     uint
	checklistItemId uint
	completed       bool
}

// GetTransactionalQueryFunction returns a transaction function to toggle item completion and update its position
func (m *toggleCompletionQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (domain.ChecklistItem, error) {
	return func(tx pool.TransactionWrapper) (domain.ChecklistItem, error) {
		// First, update the completion status
		err := tx.QueryRow(context.Background(),
			`UPDATE CHECKLIST_ITEM
			 SET CHECKLIST_ITEM_COMPLETED = @completed
			 WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @checklistItemId
			 RETURNING CHECKLIST_ITEM_ID`,
			pgx.NamedArgs{
				"checklistId":     m.checklistId,
				"checklistItemId": m.checklistItemId,
				"completed":       m.completed,
			}).Scan(&m.checklistItemId)
		if err != nil {
			return domain.ChecklistItem{}, fmt.Errorf("failed to toggle item completion: %w", err)
		}

		// Get the target position based on whether we're completing or uncompleting
		var targetOrderNumber uint
		var targetSQL string
		var sortOrder domain.SortOrder

		if m.completed {
			// When marking as completed, move to the top of completed items
			targetSQL = `SELECT MIN(ORDER_NUMBER)
						FROM CHECKLIST_ITEMS_ORDERED_VIEW
						WHERE CHECKLIST_ID = @checklistId 
						AND CHECKLIST_ITEM_COMPLETED = @completed
						AND CHECKLIST_ITEM_ID != @itemId`
			sortOrder = domain.AscSort
		} else {
			// When marking as uncompleted, move to the end of uncompleted items
			targetSQL = `SELECT COALESCE(MAX(ORDER_NUMBER), 1)
						FROM CHECKLIST_ITEMS_ORDERED_VIEW
						WHERE CHECKLIST_ID = @checklistId 
						AND CHECKLIST_ITEM_COMPLETED = @completed
						AND CHECKLIST_ITEM_ID != @itemId`
			sortOrder = domain.DescSort
		}

		err = tx.QueryRow(context.Background(), targetSQL,
			pgx.NamedArgs{
				"checklistId": m.checklistId,
				"completed":   m.completed,
				"itemId":      m.checklistItemId,
			}).Scan(&targetOrderNumber)

		// If we got a target position
		if err == nil {
			changeOrderFn := NewChangeChecklistItemOrderQueryFunction(domain.ChangeOrderRequest{
				NewOrderNumber:  targetOrderNumber,
				ChecklistId:     m.checklistId,
				ChecklistItemId: m.checklistItemId,
				SortOrder:       sortOrder, // ASC for completing (top), DESC for uncompleting (bottom)
			})

			_, err = changeOrderFn.GetTransactionalQueryFunction()(tx)
			if err != nil {
				return domain.ChecklistItem{}, fmt.Errorf("failed to update item position: %w", err)
			}
		}

		// Finally, get and return the updated item
		var item domain.ChecklistItem
		err = tx.QueryRow(context.Background(),
			`SELECT CHECKLIST_ITEM_ID, CHECKLIST_ITEM_NAME, CHECKLIST_ITEM_COMPLETED
			 FROM CHECKLIST_ITEM 
			 WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @checklistItemId`,
			pgx.NamedArgs{
				"checklistId":     m.checklistId,
				"checklistItemId": m.checklistItemId,
			}).Scan(&item.Id, &item.Name, &item.Completed)
		if err != nil {
			return domain.ChecklistItem{}, fmt.Errorf("failed to get updated item: %w", err)
		}

		return item, nil
	}
}
