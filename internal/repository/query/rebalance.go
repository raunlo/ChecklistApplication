package query

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/pool"
)

// RebalancePositionsQueryFunction redistributes positions evenly for all items in a checklist
type RebalancePositionsQueryFunction struct {
	checklistId uint
}

func NewRebalancePositionsQueryFunction(checklistId uint) *RebalancePositionsQueryFunction {
	return &RebalancePositionsQueryFunction{
		checklistId: checklistId,
	}
}

func (r *RebalancePositionsQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		// Lock all items and calculate new positions atomically
		// Maintains the order: uncompleted items first (by position), then completed items (by position)
		// Each section gets evenly-spaced positions starting from FirstItemPosition
		rebalanceSQL := `
			WITH numbered_items AS (
				SELECT
					CHECKLIST_ITEM_ID,
					CHECKLIST_ITEM_COMPLETED,
					ROW_NUMBER() OVER (
						PARTITION BY CHECKLIST_ITEM_COMPLETED
						ORDER BY POSITION
					) as row_num
				FROM CHECKLIST_ITEM
				WHERE CHECKLIST_ID = @checklistId
				FOR UPDATE
			),
			new_positions AS (
				SELECT
					CHECKLIST_ITEM_ID,
					(@startPosition + (row_num - 1) * @gap)::DOUBLE PRECISION as new_position
				FROM numbered_items
			)
			UPDATE CHECKLIST_ITEM ci
			SET POSITION = np.new_position
			FROM new_positions np
			WHERE ci.CHECKLIST_ITEM_ID = np.CHECKLIST_ITEM_ID`

		_, err := tx.Exec(context.Background(), rebalanceSQL, pgx.NamedArgs{
			"checklistId":   r.checklistId,
			"startPosition": domain.FirstItemPosition,
			"gap":           domain.DefaultGapSize,
		})

		return err == nil, err
	}
}
