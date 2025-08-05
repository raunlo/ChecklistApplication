package query

import (
	"context"
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
