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

func (u *UpdateChecklistItemRowsQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	getQuery := func(index int, row domain.ChecklistItemRow) (string, pgx.NamedArgs) {
		rowNameParamName := getIndexedSQLValueParamName(index, "rowName")
		rowParamCompletedName := getIndexedSQLValueParamName(index, "rowCompleted")
		checklistItemIdParamName := getIndexedSQLValueParamName(index, "checklistItemId")
		checklistItemRowIdParamName := getIndexedSQLValueParamName(index, "checklistItemRowId")

		sql := `UPDATE checklist_item_row 
				 SET NAME = @%s AND COMPLETED = @%s
				 WHERE ID = @%s AND CHECKLIST_ITEM_ID = @%s`
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
		tag, err := tx.SendBatch(context.Background(), batch).Exec()
		return int(tag.RowsAffected()) == len(u.checklistItemRows), err
	}
}

// UpdateChecklistItemRowsQueryFunction update checklist item rows query struct
type UpdateChecklistItemRowsQueryFunction struct {
	checklistItemId   uint
	checklistItemRows []domain.ChecklistItemRow
}

func (q *PersistChecklistItemRowQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) ([]domain.ChecklistItemRow, error) {
	return func(tx pool.TransactionWrapper) ([]domain.ChecklistItemRow, error) {
		if len(q.checklistItemRows) == 0 {
			return []domain.ChecklistItemRow{}, nil
		}
		namedArgumentsMap := pgx.NamedArgs{}
		query := "INSERT INTO checklist_item_row(ID, CHECKLIST_ITEM_ID, NAME, COMPLETED) VALUES "
		getSequenceValuesQuery := GetSequenceValuesQuery{
			sequenceName:   "checklist_item_row_id_sequence",
			numberOfValues: len(q.checklistItemRows),
		}
		ids, err := getSequenceValuesQuery.GetTransactionalQueryFunction()(tx)
		if err != nil {
			return nil, err
		}

		for index := 0; index < len(q.checklistItemRows); index++ {
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
