package query

import (
	"context"
	"errors"
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/pool"
)

// PersistChecklistItemRowQueryFunction persist checklist item rows query struct
type PersistChecklistItemRowQueryFunction struct {
	checklistItemId   uint
	checklistItemRows []domain.ChecklistItemRow
}

func (q *PersistChecklistItemRowQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) ([]domain.ChecklistItemRow, error) {
	return func(tx pool.TransactionWrapper) ([]domain.ChecklistItemRow, error) {
		if len(q.checklistItemRows) == 0 {
			return []domain.ChecklistItemRow{}, nil
		}
		namedArgumentsMap := pgx.NamedArgs{}
		query := "INSERT INTO CHECKLIST_ITEM_ROW(CHECKLIST_ITEM_ROW_ID, CHECKLIST_ITEM_ID, CHECKLIST_ITEM_ROW_NAME, CHECKLIST_ITEM_ROW_COMPLETED) VALUES "
		getSequenceValuesQuery := GetSequenceValuesQuery{
			sequenceName:   "checklist_item_row_id_sequence",
			numberOfValues: len(q.checklistItemRows),
		}
		ids, err := getSequenceValuesQuery.GetTransactionalQueryFunction()(tx)
		if err != nil {
			return nil, err
		}

		for index := range q.checklistItemRows {
			rowPointer := &q.checklistItemRows[index]
			rowPointer.Id = ids[index]

			itemRowIdParamName := getIndexedSQLValueParamName(index, "checklist_item_row_id")
			itemIdParamName := getIndexedSQLValueParamName(index, "checklist_item_id")
			itemRowNameParamName := getIndexedSQLValueParamName(index, "checklist_item_row_name")
			itemRowCompletedParamName := getIndexedSQLValueParamName(index, "checklist_item_row_completed")
			query += fmt.Sprintf(" (@%s, @%s, @%s, @%s)",
				itemRowIdParamName, itemIdParamName, itemRowNameParamName, itemRowCompletedParamName)
			if index != len(q.checklistItemRows)-1 {
				query += ", "
			} else {
				query += " "
			}

			// set parameter values
			namedArgumentsMap[itemIdParamName] = q.checklistItemId
			namedArgumentsMap[itemRowIdParamName] = rowPointer.Id
			namedArgumentsMap[itemRowNameParamName] = rowPointer.Name
			namedArgumentsMap[itemRowCompletedParamName] = rowPointer.Completed
		}
		_, err = tx.Exec(context.Background(), query, namedArgumentsMap)
		return q.checklistItemRows, err
	}
}

// UpdateChecklistItemRowsQueryFunction update checklist item rows query struct
type UpdateChecklistItemRowsQueryFunction struct {
	checklistItemId   uint
	checklistItemRows []domain.ChecklistItemRow
}

func (u *UpdateChecklistItemRowsQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	getQuery := func(index int, row domain.ChecklistItemRow) (string, pgx.NamedArgs) {
		rowNameParamName := getIndexedSQLValueParamName(index, "rowName")
		rowParamCompletedName := getIndexedSQLValueParamName(index, "rowCompleted")
		checklistItemIdParamName := getIndexedSQLValueParamName(index, "checklistItemId")
		checklistItemRowIdParamName := getIndexedSQLValueParamName(index, "checklistItemRowId")

		sql := `UPDATE CHECKLIST_ITEM_ROW 
				 SET CHECKLIST_ITEM_ROW_NAME = @%s , CHECKLIST_ITEM_ROW_COMPLETED = @%s
				 WHERE CHECKLIST_ITEM_ROW_ID = @%s AND CHECKLIST_ITEM_ID = @%s`
		sql = fmt.Sprintf(sql, rowNameParamName, rowParamCompletedName, checklistItemRowIdParamName, checklistItemIdParamName)
		args := pgx.NamedArgs{
			rowNameParamName:            row.Name,
			rowParamCompletedName:       row.Completed,
			checklistItemIdParamName:    u.checklistItemId,
			checklistItemRowIdParamName: row.Id,
		}
		return sql, args
	}
	return func(tx pool.TransactionWrapper) (bool, error) {
		batch := &pgx.Batch{}
		for index, row := range u.checklistItemRows {
			sql, args := getQuery(index, row)
			batch.Queue(sql, args)
		}
		br := tx.SendBatch(context.Background(), batch)
		defer br.Close()
		rowsAffected := 0
		var err error
		for i := 0; i < batch.Len(); i++ {
			tag, err := br.Exec()
			if err != nil {
				return false, err
			}
			fmt.Println("Rows affected:", tag.RowsAffected())
			rowsAffected += int(tag.RowsAffected())
		}
		return rowsAffected == batch.Len(), err
	}
}

// DeleteChecklistItemRowAndAutoCompleteQueryFunction delete checklist item row by id and auto-complete parent item query struct
type DeleteChecklistItemRowAndAutoCompleteQueryFunction struct {
	checklistId     uint
	checklistItemId uint
	rowId           uint
}

func (d *DeleteChecklistItemRowAndAutoCompleteQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (domain.ChecklistItemRowDeletionResult, error) {
	return func(tx pool.TransactionWrapper) (domain.ChecklistItemRowDeletionResult, error) {
		// Step 1: Lock the parent item FIRST to prevent concurrent modifications
		// This ensures no other transaction can modify the item or its rows during our operation
		lockSQL := `
			SELECT CHECKLIST_ITEM_ID 
			FROM CHECKLIST_ITEM
			WHERE CHECKLIST_ITEM_ID = @checklist_item_id
			  AND CHECKLIST_ID = @checklist_id
			FOR UPDATE`

		var itemId uint
		err := tx.QueryRow(context.Background(), lockSQL, pgx.NamedArgs{
			"checklist_item_id": d.checklistItemId,
			"checklist_id":      d.checklistId,
		}).Scan(&itemId)

		if err != nil {
			if err == pgx.ErrNoRows {
				// Parent item doesn't exist
				return domain.ChecklistItemRowDeletionResult{Success: false, ItemAutoCompleted: false}, nil
			}
			return domain.ChecklistItemRowDeletionResult{Success: false, ItemAutoCompleted: false}, err
		}

		// Step 2: Delete the row (with parent item locked, safe from concurrent changes)
		deleteSQL := `
			DELETE FROM CHECKLIST_ITEM_ROW
			WHERE CHECKLIST_ITEM_ROW_ID = @checklist_item_row_id 
			  AND CHECKLIST_ITEM_ID = @checklist_item_id`

		deleteResult, err := tx.Exec(context.Background(), deleteSQL, pgx.NamedArgs{
			"checklist_item_row_id": d.rowId,
			"checklist_item_id":     d.checklistItemId,
		})

		if err != nil {
			return domain.ChecklistItemRowDeletionResult{Success: false, ItemAutoCompleted: false}, err
		}

		if deleteResult.RowsAffected() == 0 {
			// Row not found or already deleted
			return domain.ChecklistItemRowDeletionResult{Success: false, ItemAutoCompleted: false}, nil
		}

		if deleteResult.RowsAffected() > 1 {
			return domain.ChecklistItemRowDeletionResult{Success: false, ItemAutoCompleted: false},
				errors.New("deleteChecklistItemRow affected more than one row")
		}

		// Step 3: Auto-complete parent item if all remaining rows are completed
		// This UPDATE is atomic and only succeeds if:
		// - Item is not already completed
		// - No incomplete rows remain
		// - At least one row still exists (prevents marking empty items as complete)
		autoCompleteSQL := `
			UPDATE CHECKLIST_ITEM
			SET CHECKLIST_ITEM_COMPLETED = true
			WHERE CHECKLIST_ITEM_ID = @checklist_item_id
			  AND CHECKLIST_ID = @checklist_id
			  AND CHECKLIST_ITEM_COMPLETED = false
			  AND EXISTS (
				  SELECT 1
				  FROM CHECKLIST_ITEM_ROW
				  WHERE CHECKLIST_ITEM_ID = @checklist_item_id
			  )
			  AND NOT EXISTS (
				  SELECT 1
				  FROM CHECKLIST_ITEM_ROW
				  WHERE CHECKLIST_ITEM_ID = @checklist_item_id
					AND CHECKLIST_ITEM_ROW_COMPLETED = false
			  )`

		autoCompleteResult, autoCompleteErr := tx.Exec(context.Background(), autoCompleteSQL, pgx.NamedArgs{
			"checklist_item_id": d.checklistItemId,
			"checklist_id":      d.checklistId,
		})

		if autoCompleteErr != nil {
			// Row deleted successfully but auto-complete failed
			// Still return success for deletion
			return domain.ChecklistItemRowDeletionResult{
				Success:           true,
				ItemAutoCompleted: false,
			}, autoCompleteErr
		}

		// Check if auto-completion occurred
		itemAutoCompleted := autoCompleteResult.RowsAffected() > 0

		return domain.ChecklistItemRowDeletionResult{
			Success:           true,
			ItemAutoCompleted: itemAutoCompleted,
		}, nil
	}
}
