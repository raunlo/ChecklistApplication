package query

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/pool"
)

// ChangeChecklistItemOrderQueryFunction moves item to different order number using gap-based positioning
type ChangeChecklistItemOrderQueryFunction struct {
	newOrderNumber  uint
	checklistId     uint
	checklistItemId uint
	sortOrder       domain.SortOrder
}

// positionQueryResult holds the result of position queries
type positionQueryResult struct {
	Position float64
}

func (c *ChangeChecklistItemOrderQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (domain.ChangeOrderResponse, error) {
	return func(tx pool.TransactionWrapper) (domain.ChangeOrderResponse, error) {
		// 1. Get the target item's current completed status (with lock)
		var itemCompleted bool
		err := tx.QueryRow(context.Background(),
			`SELECT CHECKLIST_ITEM_COMPLETED FROM CHECKLIST_ITEM
			 WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @itemId FOR UPDATE`,
			pgx.NamedArgs{
				"checklistId": c.checklistId,
				"itemId":      c.checklistItemId,
			}).Scan(&itemCompleted)
		if err != nil {
			return domain.ChangeOrderResponse{}, err
		}

		// 2. Calculate new position based on target order number
		newPosition, err := c.calculateNewPosition(tx, itemCompleted)
		if err != nil {
			return domain.ChangeOrderResponse{}, err
		}

		// 3. Update the item's position
		_, err = tx.Exec(context.Background(),
			`UPDATE CHECKLIST_ITEM SET POSITION = @newPosition
			 WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_ID = @itemId`,
			pgx.NamedArgs{
				"checklistId": c.checklistId,
				"itemId":      c.checklistItemId,
				"newPosition": newPosition,
			})
		if err != nil {
			return domain.ChangeOrderResponse{}, err
		}

		// 4. Check if rebalancing is needed
		rebalanceNeeded := c.checkRebalanceNeeded(tx, itemCompleted)

		return domain.ChangeOrderResponse{
			OrderNumber:     c.newOrderNumber,
			ChecklistItemId: c.checklistItemId,
			ChecklistId:     c.checklistId,
			Position:        newPosition,
			RebalanceNeeded: rebalanceNeeded,
		}, nil
	}
}

func (c *ChangeChecklistItemOrderQueryFunction) calculateNewPosition(tx pool.TransactionWrapper, completed bool) (float64, error) {
	// Get ordered positions in the same completion section, excluding the moving item
	rows, err := tx.Query(context.Background(),
		`SELECT POSITION FROM CHECKLIST_ITEM
		 WHERE CHECKLIST_ID = @checklistId
		   AND CHECKLIST_ITEM_COMPLETED = @completed
		   AND CHECKLIST_ITEM_ID != @itemId
		 ORDER BY POSITION ASC`,
		pgx.NamedArgs{
			"checklistId": c.checklistId,
			"completed":   completed,
			"itemId":      c.checklistItemId,
		})
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var positions []float64
	for rows.Next() {
		var result positionQueryResult
		if err := rows.Scan(&result.Position); err != nil {
			return 0, err
		}
		positions = append(positions, result.Position)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	// Target order number is 1-based
	targetIndex := int(c.newOrderNumber) - 1

	// Calculate new position based on target index
	if len(positions) == 0 {
		// No other items, use default position
		return domain.FirstItemPosition, nil
	}

	if targetIndex <= 0 {
		// Insert at the beginning
		return positions[0] - domain.DefaultGapSize, nil
	}

	if targetIndex >= len(positions) {
		// Insert at the end
		return positions[len(positions)-1] + domain.DefaultGapSize, nil
	}

	// Insert between two items
	prevPosition := positions[targetIndex-1]
	nextPosition := positions[targetIndex]
	return (prevPosition + nextPosition) / 2, nil
}

func (c *ChangeChecklistItemOrderQueryFunction) checkRebalanceNeeded(tx pool.TransactionWrapper, completed bool) bool {
	// Check if any adjacent gap is too small
	var minGap float64
	err := tx.QueryRow(context.Background(),
		`WITH positions AS (
			SELECT POSITION,
				   LAG(POSITION) OVER (ORDER BY POSITION) as prev_pos
			FROM CHECKLIST_ITEM
			WHERE CHECKLIST_ID = @checklistId AND CHECKLIST_ITEM_COMPLETED = @completed
		)
		SELECT COALESCE(MIN(POSITION - prev_pos), @defaultGap)
		FROM positions WHERE prev_pos IS NOT NULL`,
		pgx.NamedArgs{
			"checklistId": c.checklistId,
			"completed":   completed,
			"defaultGap":  domain.DefaultGapSize,
		}).Scan(&minGap)

	if err != nil {
		return false
	}

	return minGap < domain.MinGapThreshold
}
