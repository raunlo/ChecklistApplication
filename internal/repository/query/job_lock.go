package query

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/raunlo/pgx-with-automapper/pool"
)

const jobNameSoftDeleteCleanup = "soft_delete_cleanup"

// TryAcquireCleanupLockQueryFunction attempts to acquire the cleanup job lock
// using PostgreSQL's SELECT FOR UPDATE SKIP LOCKED to prevent concurrent execution.
// Returns true if:
// 1. The lock was successfully acquired (no other instance holds it)
// 2. Enough time has passed since the last successful run
type TryAcquireCleanupLockQueryFunction struct {
	minInterval time.Duration
}

func NewTryAcquireCleanupLockQueryFunction(minInterval time.Duration) *TryAcquireCleanupLockQueryFunction {
	return &TryAcquireCleanupLockQueryFunction{minInterval: minInterval}
}

func (q *TryAcquireCleanupLockQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		// Try to acquire the lock using SKIP LOCKED (non-blocking)
		// This returns no rows if another instance holds the lock
		var lastRunAt time.Time
		err := tx.QueryRow(context.Background(), `
			SELECT last_run_at 
			FROM job_lock 
			WHERE job_name = $1 
			  AND (locked_by IS NULL OR locked_at < NOW() - INTERVAL '10 minutes')
			FOR UPDATE SKIP LOCKED
		`, jobNameSoftDeleteCleanup).Scan(&lastRunAt)

		if err == pgx.ErrNoRows {
			// Lock is held by another instance - not an error, just means we skip
			return false, nil
		}
		if err != nil {
			return false, err
		}

		// Check if enough time has passed since last run
		if time.Since(lastRunAt) < q.minInterval {
			// Not enough time passed - not an error, just means we skip
			return false, nil
		}

		// Acquire the lock by setting locked_by and locked_at
		instanceID := generateInstanceID()
		_, err = tx.Exec(context.Background(), `
			UPDATE job_lock 
			SET locked_by = $1, locked_at = NOW() 
			WHERE job_name = $2
		`, instanceID, jobNameSoftDeleteCleanup)
		if err != nil {
			return false, err
		}

		return true, nil
	}
}

// ReleaseCleanupLockQueryFunction releases the cleanup job lock
type ReleaseCleanupLockQueryFunction struct{}

func NewReleaseCleanupLockQueryFunction() *ReleaseCleanupLockQueryFunction {
	return &ReleaseCleanupLockQueryFunction{}
}

func (q *ReleaseCleanupLockQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		_, err := tx.Exec(context.Background(), `
			UPDATE job_lock 
			SET locked_by = NULL, locked_at = NULL 
			WHERE job_name = $1
		`, jobNameSoftDeleteCleanup)
		return err == nil, err
	}
}

// UpdateCleanupLastRunQueryFunction updates the last run timestamp and releases the lock
type UpdateCleanupLastRunQueryFunction struct{}

func NewUpdateCleanupLastRunQueryFunction() *UpdateCleanupLastRunQueryFunction {
	return &UpdateCleanupLastRunQueryFunction{}
}

func (q *UpdateCleanupLastRunQueryFunction) GetTransactionalQueryFunction() func(tx pool.TransactionWrapper) (bool, error) {
	return func(tx pool.TransactionWrapper) (bool, error) {
		_, err := tx.Exec(context.Background(), `
			UPDATE job_lock 
			SET last_run_at = NOW(), locked_by = NULL, locked_at = NULL 
			WHERE job_name = $1
		`, jobNameSoftDeleteCleanup)
		return err == nil, err
	}
}

// generateInstanceID creates a unique identifier for this instance
func generateInstanceID() string {
	return time.Now().Format("20060102150405.000000")
}
