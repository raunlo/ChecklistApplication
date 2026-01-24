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
		// Calculate target position based on completion status
		var newPosition float64
		var positionQuery string

		if m.completed {
			// Move to beginning of completed section (smallest position in completed, or default if none)
			positionQuery = `SELECT COALESCE(MIN(POSITION) - @gap, @defaultPos)
							 FROM CHECKLIST_ITEM
							 WHERE CHECKLIST_ID = @checklistId
							   AND CHECKLIST_ITEM_COMPLETED = TRUE
							   AND CHECKLIST_ITEM_ID != @itemId`
		} else {
			// Move to end of uncompleted section (largest position in uncompleted, or default if none)
			positionQuery = `SELECT COALESCE(MAX(POSITION) + @gap, @defaultPos)
							 FROM CHECKLIST_ITEM
							 WHERE CHECKLIST_ID = @checklistId
							   AND CHECKLIST_ITEM_COMPLETED = FALSE
							   AND CHECKLIST_ITEM_ID != @itemId`
		}

		err := tx.QueryRow(context.Background(), positionQuery, pgx.NamedArgs{
			"checklistId": m.checklistId,
			"itemId":      m.checklistItemId,
			"gap":         domain.DefaultGapSize,
			"defaultPos":  domain.FirstItemPosition,
		}).Scan(&newPosition)
		if err != nil {
			return domain.ChecklistItem{}, fmt.Errorf("failed to calculate new position: %w", err)
		}

		// Update completion status and position atomically
		var item domain.ChecklistItem
		err = tx.QueryRow(context.Background(),
			`UPDATE CHECKLIST_ITEM
			 SET CHECKLIST_ITEM_COMPLETED = @completed, POSITION = @newPosition
			 WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @checklistItemId
			 RETURNING CHECKLIST_ITEM_ID, CHECKLIST_ITEM_NAME, CHECKLIST_ITEM_COMPLETED, POSITION`,
			pgx.NamedArgs{
				"checklistId":     m.checklistId,
				"checklistItemId": m.checklistItemId,
				"completed":       m.completed,
				"newPosition":     newPosition,
			}).Scan(&item.Id, &item.Name, &item.Completed, &item.Position)
		if err != nil {
			return domain.ChecklistItem{}, fmt.Errorf("failed to toggle item completion: %w", err)
		}

		return item, nil
	}
}
