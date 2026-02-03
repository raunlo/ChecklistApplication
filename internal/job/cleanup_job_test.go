package job

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"com.raunlo.checklist/internal/core/domain"
)

// mockRepository implements the IChecklistItemsRepository interface for testing
type mockRepository struct {
	purgeCallCount       atomic.Int32
	purgeReturn          int64
	purgeError           domain.Error
	tryAcquireLockReturn bool
	tryAcquireLockError  domain.Error
	acquireLockCallCount atomic.Int32
}

func (m *mockRepository) PurgeSoftDeletedItems(ctx context.Context, retentionPeriod time.Duration) (int64, domain.Error) {
	m.purgeCallCount.Add(1)
	return m.purgeReturn, m.purgeError
}

func (m *mockRepository) TryAcquireCleanupLock(ctx context.Context, minInterval time.Duration) (bool, domain.Error) {
	m.acquireLockCallCount.Add(1)
	return m.tryAcquireLockReturn, m.tryAcquireLockError
}

func (m *mockRepository) ReleaseCleanupLock(ctx context.Context) domain.Error {
	return nil
}

func (m *mockRepository) UpdateCleanupLastRun(ctx context.Context) domain.Error {
	return nil
}

// Implement other required interface methods (not used in cleanup job)
func (m *mockRepository) UpdateChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	return domain.ChecklistItem{}, nil
}
func (m *mockRepository) SaveChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	return domain.ChecklistItem{}, nil
}
func (m *mockRepository) SaveChecklistItemRow(ctx context.Context, checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error) {
	return domain.ChecklistItemRow{}, nil
}
func (m *mockRepository) FindChecklistItemById(ctx context.Context, checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
	return nil, nil
}
func (m *mockRepository) DeleteChecklistItemById(ctx context.Context, checklistId uint, id uint) domain.Error {
	return nil
}
func (m *mockRepository) DeleteChecklistItemRowAndAutoComplete(ctx context.Context, checklistId uint, itemId uint, rowId uint) (domain.ChecklistItemRowDeletionResult, domain.Error) {
	return domain.ChecklistItemRowDeletionResult{}, nil
}
func (m *mockRepository) FindAllChecklistItems(ctx context.Context, checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error) {
	return nil, nil
}
func (m *mockRepository) ChangeChecklistItemOrder(ctx context.Context, request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error) {
	return domain.ChangeOrderResponse{}, nil
}
func (m *mockRepository) ToggleItemCompleted(ctx context.Context, checklistId uint, checklistItemId uint, completed bool) (domain.ChecklistItem, domain.Error) {
	return domain.ChecklistItem{}, nil
}
func (m *mockRepository) RebalancePositions(ctx context.Context, checklistId uint) domain.Error {
	return nil
}
func (m *mockRepository) RestoreChecklistItem(ctx context.Context, checklistId uint, itemId uint) (domain.ChecklistItem, domain.Error) {
	return domain.ChecklistItem{}, nil
}

func TestCleanupJob_RunsOnStartAndPeriodically(t *testing.T) {
	repo := &mockRepository{
		purgeReturn:          5,
		tryAcquireLockReturn: true, // Lock acquired successfully
	}

	config := CleanupJobConfig{
		RetentionPeriod: 30 * 24 * time.Hour,
		Interval:        100 * time.Millisecond, // Short interval for testing
	}

	job := NewCleanupJob(repo, config)
	job.Start()

	// Wait for initial run + at least one periodic run
	time.Sleep(250 * time.Millisecond)
	job.Stop()

	// Should have been called at least 2 times (initial + periodic)
	callCount := int(repo.purgeCallCount.Load())
	if callCount < 2 {
		t.Errorf("expected at least 2 purge calls, got %d", callCount)
	}
}

func TestCleanupJob_SkipsWhenLockNotAcquired(t *testing.T) {
	repo := &mockRepository{
		purgeReturn:          5,
		tryAcquireLockReturn: false, // Lock not acquired (another instance is running or ran recently)
	}

	config := CleanupJobConfig{
		RetentionPeriod: 30 * 24 * time.Hour,
		Interval:        100 * time.Millisecond,
	}

	job := NewCleanupJob(repo, config)
	job.Start()

	time.Sleep(250 * time.Millisecond)
	job.Stop()

	// Lock should have been attempted
	lockAttempts := int(repo.acquireLockCallCount.Load())
	if lockAttempts < 2 {
		t.Errorf("expected at least 2 lock attempts, got %d", lockAttempts)
	}

	// Purge should NOT have been called since lock wasn't acquired
	callCount := int(repo.purgeCallCount.Load())
	if callCount != 0 {
		t.Errorf("expected 0 purge calls when lock not acquired, got %d", callCount)
	}
}

func TestCleanupJob_DefaultConfig(t *testing.T) {
	config := DefaultCleanupJobConfig()

	if config.RetentionPeriod != 30*24*time.Hour {
		t.Errorf("expected retention period of 30 days, got %v", config.RetentionPeriod)
	}

	if config.Interval != 24*time.Hour {
		t.Errorf("expected interval of 24 hours, got %v", config.Interval)
	}
}

func TestNewCleanupJob_DefaultsZeroValues(t *testing.T) {
	repo := &mockRepository{tryAcquireLockReturn: true}

	// Pass zero values
	job := NewCleanupJob(repo, CleanupJobConfig{})

	if job.retentionPeriod != 30*24*time.Hour {
		t.Errorf("expected default retention period of 30 days, got %v", job.retentionPeriod)
	}

	if job.interval != 24*time.Hour {
		t.Errorf("expected default interval of 24 hours, got %v", job.interval)
	}
}
