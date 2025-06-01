package query

import (
	"context"
	"fmt"
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

	var query string
	for index := 0; index < q.numberOfValues; index++ {

		if index != 0 {
			query += "UNION ALL\n"
		}
		query += fmt.Sprintf("SELECT NEXTVAL('%s') as id\n", q.sequenceName)
	}
	return func(tx pgx.Tx) ([]uint, error) {
		var ids []uint
		rows, err := tx.Query(context.Background(), query)
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
