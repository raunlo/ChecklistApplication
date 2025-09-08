package query

import (
	"context"
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type PersistChecklistItemTemplateRowQueryFunction struct {
	checklistItemTemplateId   uint
	checklistItemTemplateRows []domain.ChecklistItemTemplateRow
}

func (p *PersistChecklistItemTemplateRowQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) ([]domain.ChecklistItemTemplateRow, error) {
	return func(tx pool.TransactionWrapper) ([]domain.ChecklistItemTemplateRow, error) {
		if len(p.checklistItemTemplateRows) == 0 {
			return []domain.ChecklistItemTemplateRow{}, nil
		}
		namedArgs := pgx.NamedArgs{}
		query := "INSERT INTO TEMPLATE_ITEM_ROW(ID, TEMPLATE_ITEM_ID, NAME) VALUES "

		ids, err := (&GetSequenceValuesQuery{sequenceName: "template_item_row_id", numberOfValues: len(p.checklistItemTemplateRows)}).GetTransactionalQueryFunction()(tx)
		if err != nil {
			return nil, err
		}
		for i := range p.checklistItemTemplateRows {
			rowPtr := &p.checklistItemTemplateRows[i]
			rowPtr.Id = ids[i]
			idParam := getIndexedSQLValueParamName(i, "rowId")
			templateIdParam := getIndexedSQLValueParamName(i, "templateId")
			nameParam := getIndexedSQLValueParamName(i, "rowName")
			query += fmt.Sprintf("(@%s, @%s, @%s)", idParam, templateIdParam, nameParam)
			if i != len(p.checklistItemTemplateRows)-1 {
				query += ", "
			}
			namedArgs[idParam] = rowPtr.Id
			namedArgs[templateIdParam] = p.checklistItemTemplateId
			namedArgs[nameParam] = "__template_row__"
		}
		_, err = tx.Exec(context.Background(), query, namedArgs)
		return p.checklistItemTemplateRows, err
	}

}

type UpdateChecklistItemTemplateRowsQueryFunction struct {
	checklistItemTemplateId   uint
	checklistItemTemplateRows []domain.ChecklistItemTemplateRow
}

func (u *UpdateChecklistItemTemplateRowsQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	getQuery := func(index int, row domain.ChecklistItemTemplateRow) (string, pgx.NamedArgs) {
		rowNameParam := getIndexedSQLValueParamName(index, "rowName")
		templateIdParam := getIndexedSQLValueParamName(index, "templateId")
		rowIdParam := getIndexedSQLValueParamName(index, "rowId")
		sql := fmt.Sprintf(`UPDATE TEMPLATE_ITEM_ROW SET NAME = @%s WHERE ID = @%s AND TEMPLATE_ITEM_ID = @%s`, rowNameParam, rowIdParam, templateIdParam)
		args := pgx.NamedArgs{
			rowNameParam:    "__template_row__",
			templateIdParam: u.checklistItemTemplateId,
			rowIdParam:      row.Id,
		}
		return sql, args
	}
	return func(tx pool.TransactionWrapper) (bool, error) {
		if len(u.checklistItemTemplateRows) == 0 {
			return true, nil
		}
		batch := &pgx.Batch{}
		for i, row := range u.checklistItemTemplateRows {
			sql, args := getQuery(i, row)
			batch.Queue(sql, args)
		}
		br := tx.SendBatch(context.Background(), batch)
		defer br.Close()
		rowsAffected := 0
		for i := 0; i < batch.Len(); i++ {
			tag, err := br.Exec()
			if err != nil {
				return false, err
			}
			rowsAffected += int(tag.RowsAffected())
		}
		return rowsAffected == batch.Len(), nil
	}
}
