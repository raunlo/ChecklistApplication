package job

import (
	"context"
	"log"
	"time"

	"com.raunlo.checklist/internal/core/repository"
)

// CleanupJob handles periodic cleanup of soft-deleted items.
// It is designed to work in serverless/multi-instance environments like Cloud Run by:
// 1. Using database-level locking to prevent concurrent runs
// 2. Tracking last run time in the database to ensure daily runs even across restarts
type CleanupJob struct {
	repo            repository.IChecklistItemsRepository
	retentionPeriod time.Duration
	interval        time.Duration
	stopCh          chan struct{}
}

// CleanupJobConfig holds configuration for the cleanup job
type CleanupJobConfig struct {
	// RetentionPeriod is how long soft-deleted items are kept before permanent deletion
	// Default: 30 days
	RetentionPeriod time.Duration
	// Interval is how often the cleanup job checks if it should run
	// In serverless environments, this is checked on startup
	// Default: 24 hours (daily)
	Interval time.Duration
}

// DefaultCleanupJobConfig returns the default configuration
func DefaultCleanupJobConfig() CleanupJobConfig {
	return CleanupJobConfig{
		RetentionPeriod: 30 * 24 * time.Hour, // 30 days
		Interval:        24 * time.Hour,      // Daily
	}
}

// NewCleanupJob creates a new cleanup job
func NewCleanupJob(repo repository.IChecklistItemsRepository, config CleanupJobConfig) *CleanupJob {
	if config.RetentionPeriod == 0 {
		config.RetentionPeriod = 30 * 24 * time.Hour
	}
	if config.Interval == 0 {
		config.Interval = 24 * time.Hour
	}

	return &CleanupJob{
		repo:            repo,
		retentionPeriod: config.RetentionPeriod,
		interval:        config.Interval,
		stopCh:          make(chan struct{}),
	}
}

// Start begins the cleanup job in a goroutine.
// The job uses database-level coordination to ensure:
// - Only one instance runs the cleanup at a time (prevents race conditions)
// - Cleanup runs at most once per interval, even across instance restarts
func (j *CleanupJob) Start() {
	go j.run()
	log.Printf("Cleanup job started: will purge items deleted more than %v ago, running every %v", j.retentionPeriod, j.interval)
}

// Stop gracefully stops the cleanup job
func (j *CleanupJob) Stop() {
	close(j.stopCh)
	log.Println("Cleanup job stopped")
}

func (j *CleanupJob) run() {
	// Try to run immediately on startup (will check last run time in DB)
	j.tryRunCleanup()

	// In serverless environments like Cloud Run, the instance may be short-lived.
	// We still use a ticker for long-running instances, but the actual run decision
	// is coordinated via the database.
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.tryRunCleanup()
		case <-j.stopCh:
			return
		}
	}
}

// tryRunCleanup attempts to acquire a lock and run cleanup if enough time has passed
func (j *CleanupJob) tryRunCleanup() {
	// Use a timeout to prevent hanging if database is slow/unresponsive
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Try to acquire lock and check if we should run
	shouldRun, err := j.repo.TryAcquireCleanupLock(ctx, j.interval)
	if err != nil {
		log.Printf("Cleanup job: failed to check lock: %v", err)
		return
	}

	if !shouldRun {
		log.Printf("Cleanup job: skipping - another instance ran recently or is currently running")
		return
	}

	// We have the lock and should run
	deletedCount, err := j.repo.PurgeSoftDeletedItems(ctx, j.retentionPeriod)
	if err != nil {
		log.Printf("Cleanup job error: failed to purge soft-deleted items: %v", err)
		// Release lock on error so another instance can try
		_ = j.repo.ReleaseCleanupLock(ctx)
		return
	}

	// Update last run time and release lock
	if err := j.repo.UpdateCleanupLastRun(ctx); err != nil {
		log.Printf("Cleanup job: failed to update last run time: %v", err)
		// Still release the lock so another instance can try
		_ = j.repo.ReleaseCleanupLock(ctx)
		return
	}

	if deletedCount > 0 {
		log.Printf("Cleanup job: permanently deleted %d items that were soft-deleted more than %v ago", deletedCount, j.retentionPeriod)
	} else {
		log.Printf("Cleanup job: no items to purge")
	}
}
