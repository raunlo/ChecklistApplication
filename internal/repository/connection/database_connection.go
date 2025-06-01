package connection

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type TransactionProps[QueryResultT any] struct {
	TxOptions  pgx.TxOptions
	Query      RunQueryInTransaction[QueryResultT]
	Connection pool.Conn
}

type RunQueryInTransaction[QueryResultT any] func(tx pool.TransactionWrapper) (QueryResultT, error)

func RunInTransaction[QueryResultT any](props TransactionProps[QueryResultT]) (QueryResultT, error) {
	var result QueryResultT
	if props.Connection == nil {
		return result, errors.New("DB is missing in transaction")
	}

	tx, err := props.Connection.BeginTx(context.TODO(), props.TxOptions)
	if err == nil {
		result, err = props.Query(tx)
	}

	defer func() {
		if err != nil {
			err = tx.Rollback(context.Background())
		} else {
			err = tx.Commit(context.Background())
		}
	}()
	return result, err
}
