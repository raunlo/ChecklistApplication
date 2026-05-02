package query

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type GetSequenceValuesQuery struct {
	sequenceName   string
	numberOfValues int
}

func (q *GetSequenceValuesQuery) GetTransactionalQueryFunction() func(tx pgx.Tx) ([]uint, error) {
	if q.numberOfValues < 1 {
		return nil
	}

	var query strings.Builder
	for index := range q.numberOfValues {

		if index != 0 {
			query.WriteString("UNION ALL\n")
		}
		query.WriteString(fmt.Sprintf("SELECT NEXTVAL('%s') as id\n", q.sequenceName))
	}
	return func(tx pgx.Tx) ([]uint, error) {
		var ids []uint
		rows, err := tx.Query(context.Background(), query.String())
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var id uint
			err = rows.Scan(&id)
			ids = append(ids, id)
		}
		return ids, err
	}
}
