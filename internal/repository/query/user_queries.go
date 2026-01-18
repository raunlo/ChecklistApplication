package query

import (
	"context"
	"fmt"
	"time"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/raunlo/pgx-with-automapper/pool"
)

const DeleteAllUserChecklists = `
	DELETE FROM CHECKLIST WHERE OWNER = $1
`

func GetUserDataExport(ctx context.Context, conn pool.Conn, userId string) (*domain.UserDataExport, error) {
	// Query all checklists owned by the user
	checklistRows, err := conn.Query(ctx, `
		SELECT ID, NAME
		FROM CHECKLIST
		WHERE OWNER = $1
		ORDER BY ID
	`, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query checklists: %w", err)
	}
	defer checklistRows.Close()

	var checklists []domain.ExportedChecklist

	for checklistRows.Next() {
		var checklist domain.ExportedChecklist
		if err := checklistRows.Scan(&checklist.Id, &checklist.Name); err != nil {
			return nil, fmt.Errorf("failed to scan checklist: %w", err)
		}
		checklist.CreatedAt = time.Now() // Placeholder, add created_at column if needed

		// Query items for this checklist
		itemRows, err := conn.Query(ctx, `
			SELECT
				v.CHECKLIST_ITEM_ID,
				v.CHECKLIST_ITEM_NAME,
				v.CHECKLIST_ITEM_COMPLETED,
				v.ORDER_NUMBER
			FROM CHECKLIST_ITEMS_ORDERED_VIEW v
			WHERE v.CHECKLIST_ID = $1
			ORDER BY v.ORDER_NUMBER
		`, checklist.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to query items: %w", err)
		}

		var items []domain.ExportedChecklistItem
		for itemRows.Next() {
			var item domain.ExportedChecklistItem
			if err := itemRows.Scan(&item.Id, &item.Name, &item.Completed, &item.OrderNumber); err != nil {
				itemRows.Close()
				return nil, fmt.Errorf("failed to scan item: %w", err)
			}

			// Query rows for this item
			rowRows, err := conn.Query(ctx, `
				SELECT
					CHECKLIST_ITEM_ROW_ID,
					CHECKLIST_ITEM_ROW_NAME,
					CHECKLIST_ITEM_ROW_COMPLETED
				FROM CHECKLIST_ITEM_ROW
				WHERE CHECKLIST_ITEM_ID = $1
				ORDER BY CHECKLIST_ITEM_ROW_ID
			`, item.Id)
			if err != nil {
				itemRows.Close()
				return nil, fmt.Errorf("failed to query rows: %w", err)
			}

			var rows []domain.ExportedChecklistItemRow
			for rowRows.Next() {
				var row domain.ExportedChecklistItemRow
				if err := rowRows.Scan(&row.Id, &row.Name, &row.Completed); err != nil {
					rowRows.Close()
					itemRows.Close()
					return nil, fmt.Errorf("failed to scan row: %w", err)
				}
				rows = append(rows, row)
			}
			rowRows.Close()

			item.Rows = rows
			items = append(items, item)
		}
		itemRows.Close()

		// Query shares for this checklist
		shareRows, err := conn.Query(ctx, `
			SELECT
				SHARED_WITH_USER_ID,
				PERMISSION_LEVEL,
				CREATED_AT
			FROM CHECKLIST_SHARE
			WHERE CHECKLIST_ID = $1
			ORDER BY CREATED_AT
		`, checklist.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to query shares: %w", err)
		}

		var shares []domain.ExportedChecklistShare
		for shareRows.Next() {
			var share domain.ExportedChecklistShare
			if err := shareRows.Scan(&share.SharedWithUserId, &share.PermissionLevel, &share.SharedAt); err != nil {
				shareRows.Close()
				return nil, fmt.Errorf("failed to scan share: %w", err)
			}
			shares = append(shares, share)
		}
		shareRows.Close()

		checklist.Items = items
		checklist.Shares = shares
		checklists = append(checklists, checklist)
	}

	if err := checklistRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating checklists: %w", err)
	}

	return &domain.UserDataExport{
		UserId:     userId,
		ExportedAt: time.Now(),
		Checklists: checklists,
	}, nil
}
