package connection

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/raunlo/pgx-with-automapper/pool"
)

// Predefined TxOptions for tiered isolation levels
var (
	// TxSerializable is for operations that require strict consistency:
	// - Claiming invites (prevents double-claiming)
	// - Reordering items (prevents concurrent reorder conflicts)
	// - Multi-row atomic operations
	TxSerializable = pgx.TxOptions{IsoLevel: pgx.Serializable}

	// TxReadCommitted is for simple CRUD operations:
	// - Single-row inserts, updates, deletes
	// - Operations where eventual consistency is acceptable
	TxReadCommitted = pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
)

// Retry configuration for serialization failures
const (
	maxRetries     = 3
	baseRetryDelay = 10 * time.Millisecond
)

type TransactionProps[QueryResultT any] struct {
	TxOptions  pgx.TxOptions
	Query      RunQueryInTransaction[QueryResultT]
	Connection pool.Conn
}

type RunQueryInTransaction[QueryResultT any] func(tx pool.TransactionWrapper) (QueryResultT, error)

// isSerializationError checks if the error is a PostgreSQL serialization failure
func isSerializationError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// PostgreSQL serialization failure codes: 40001 (serialization_failure), 40P01 (deadlock_detected)
	return strings.Contains(errStr, "40001") ||
		strings.Contains(errStr, "40P01") ||
		strings.Contains(errStr, "could not serialize access") ||
		strings.Contains(errStr, "deadlock detected")
}

// runSingleTransaction executes a single transaction attempt
func runSingleTransaction[QueryResultT any](props TransactionProps[QueryResultT]) (QueryResultT, error) {
	var result QueryResultT

	tx, err := props.Connection.BeginTx(context.TODO(), props.TxOptions)
	if err != nil {
		return result, err
	}

	// Ensure we always rollback or commit
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(context.Background())
		}
	}()

	result, err = props.Query(tx)
	if err != nil {
		return result, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return result, err
	}
	committed = true

	return result, nil
}

func RunInTransaction[QueryResultT any](props TransactionProps[QueryResultT]) (QueryResultT, error) {
	var result QueryResultT
	if props.Connection == nil {
		return result, errors.New("DB is missing in transaction")
	}

	// Only retry for SERIALIZABLE isolation level
	shouldRetry := props.TxOptions.IsoLevel == pgx.Serializable

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, lastErr = runSingleTransaction(props)

		if lastErr == nil {
			return result, nil
		}

		// Only retry on serialization errors for SERIALIZABLE transactions
		if !shouldRetry || !isSerializationError(lastErr) {
			return result, lastErr
		}

		// Don't sleep after the last attempt
		if attempt < maxRetries {
			// Exponential backoff: 10ms, 20ms, 40ms
			delay := baseRetryDelay * time.Duration(1<<attempt)
			log.Printf("Serialization failure (attempt %d/%d), retrying in %v: %v",
				attempt+1, maxRetries+1, delay, lastErr)
			time.Sleep(delay)
		}
	}

	log.Printf("Transaction failed after %d attempts: %v", maxRetries+1, lastErr)
	return result, lastErr
}
