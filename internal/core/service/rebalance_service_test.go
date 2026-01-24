package service

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"com.raunlo.checklist/internal/core/domain"
)

// rebalanceMockRepo wraps mockChecklistItemsRepository to track rebalance calls
type rebalanceMockRepo struct {
	*mockChecklistItemsRepository
	rebalanceCalls int32
	callDelay      time.Duration
	mu             sync.Mutex
	callHistory    []uint
}

func newRebalanceMockRepo(delay time.Duration) *rebalanceMockRepo {
	return &rebalanceMockRepo{
		mockChecklistItemsRepository: new(mockChecklistItemsRepository),
		callDelay:                    delay,
		callHistory:                  []uint{},
	}
}

func (r *rebalanceMockRepo) RebalancePositions(ctx context.Context, checklistId uint) domain.Error {
	atomic.AddInt32(&r.rebalanceCalls, 1)
	// Simulate some work
	if r.callDelay > 0 {
		time.Sleep(r.callDelay)
	}

	r.mu.Lock()
	r.callHistory = append(r.callHistory, checklistId)
	r.mu.Unlock()

	return nil
}

func (r *rebalanceMockRepo) GetRebalanceCalls() int32 {
	return atomic.LoadInt32(&r.rebalanceCalls)
}

func (r *rebalanceMockRepo) GetCallHistory() []uint {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]uint{}, r.callHistory...)
}

func TestRebalanceService_SingleTrigger(t *testing.T) {
	repo := newRebalanceMockRepo(0)
	service := NewRebalanceService(repo)

	service.TriggerRebalance(1)

	// Wait for debounce to fire
	time.Sleep(600 * time.Millisecond)

	calls := repo.GetRebalanceCalls()
	if calls != 1 {
		t.Errorf("Expected 1 rebalance call, got %d", calls)
	}
}

func TestRebalanceService_DebounceMultipleTriggers(t *testing.T) {
	repo := newRebalanceMockRepo(0)
	service := NewRebalanceService(repo)

	// Trigger 10 times rapidly (within debounce window)
	for i := 0; i < 10; i++ {
		service.TriggerRebalance(1)
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for final debounce to fire
	time.Sleep(600 * time.Millisecond)

	calls := repo.GetRebalanceCalls()
	if calls != 1 {
		t.Errorf("Expected 1 rebalance call (debounced), got %d", calls)
	}
}

func TestRebalanceService_MaxWaitTimer(t *testing.T) {
	repo := newRebalanceMockRepo(0)
	service := NewRebalanceService(repo)

	// Trigger every 400ms for 6 seconds (should hit maxWait at 5s)
	done := make(chan bool)
	go func() {
		for i := 0; i < 15; i++ {
			service.TriggerRebalance(1)
			time.Sleep(400 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for maxWait to trigger (should be ~5 seconds)
	time.Sleep(5500 * time.Millisecond)

	calls := repo.GetRebalanceCalls()
	if calls < 1 {
		t.Errorf("Expected at least 1 rebalance call (maxWait), got %d", calls)
	}

	<-done
}

func TestRebalanceService_ConcurrentTriggersHighLoad(t *testing.T) {
	repo := newRebalanceMockRepo(10 * time.Millisecond)
	service := NewRebalanceService(repo)

	var wg sync.WaitGroup
	numGoroutines := 100
	triggersPerGoroutine := 10

	// Launch 100 goroutines, each triggering 10 times
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < triggersPerGoroutine; j++ {
				service.TriggerRebalance(1)
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Wait for debounce/maxWait to complete
	time.Sleep(6 * time.Second)

	calls := repo.GetRebalanceCalls()
	t.Logf("Total triggers: %d, Actual rebalance calls: %d", numGoroutines*triggersPerGoroutine, calls)

	// Should have much fewer calls than triggers (debouncing worked)
	if calls > 10 {
		t.Errorf("Expected debouncing to reduce calls to <10, got %d", calls)
	}

	if calls < 1 {
		t.Errorf("Expected at least 1 rebalance call, got %d", calls)
	}
}

func TestRebalanceService_MultipleChecklists(t *testing.T) {
	repo := newRebalanceMockRepo(0)
	service := NewRebalanceService(repo)

	// Trigger 3 different checklists
	service.TriggerRebalance(1)
	service.TriggerRebalance(2)
	service.TriggerRebalance(3)

	// Wait for debounces to fire
	time.Sleep(600 * time.Millisecond)

	calls := repo.GetRebalanceCalls()
	if calls != 3 {
		t.Errorf("Expected 3 rebalance calls (one per checklist), got %d", calls)
	}

	history := repo.GetCallHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 entries in history, got %d", len(history))
	}
}

func TestRebalanceService_NoDoubleExecution(t *testing.T) {
	repo := newRebalanceMockRepo(50 * time.Millisecond)
	service := NewRebalanceService(repo)

	// Trigger once
	service.TriggerRebalance(1)

	// Wait for debounce
	time.Sleep(600 * time.Millisecond)

	// Trigger again immediately after first execution
	service.TriggerRebalance(1)

	// Wait for second debounce
	time.Sleep(600 * time.Millisecond)

	calls := repo.GetRebalanceCalls()
	if calls != 2 {
		t.Errorf("Expected exactly 2 rebalance calls, got %d", calls)
	}
}

func TestRebalanceService_RaceCondition(t *testing.T) {
	// This test tries to trigger the race condition where executeOnce
	// and TriggerRebalance happen at the same time
	repo := newRebalanceMockRepo(100 * time.Millisecond)
	service := NewRebalanceService(repo)

	var wg sync.WaitGroup

	// Goroutine 1: Trigger continuously
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			service.TriggerRebalance(1)
			time.Sleep(20 * time.Millisecond)
		}
	}()

	// Goroutine 2: Also trigger continuously (creates contention)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			service.TriggerRebalance(1)
			time.Sleep(25 * time.Millisecond)
		}
	}()

	wg.Wait()

	// Wait for all pending rebalances
	time.Sleep(6 * time.Second)

	calls := repo.GetRebalanceCalls()
	t.Logf("Race condition test: %d rebalance calls from 100 triggers", calls)

	// Should be debounced to a small number
	if calls > 15 {
		t.Errorf("Too many rebalance calls, debouncing may not be working correctly: %d", calls)
	}

	if calls < 1 {
		t.Errorf("Expected at least 1 rebalance call, got %d", calls)
	}
}

func TestRebalanceService_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	repo := newRebalanceMockRepo(5 * time.Millisecond)
	service := NewRebalanceService(repo)

	var wg sync.WaitGroup
	numGoroutines := 500
	triggersPerGoroutine := 20
	totalTriggers := numGoroutines * triggersPerGoroutine

	startTime := time.Now()

	// Launch many goroutines triggering many times
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			checklistID := uint(id%10 + 1) // Use 10 different checklists
			for j := 0; j < triggersPerGoroutine; j++ {
				service.TriggerRebalance(checklistID)
				time.Sleep(time.Duration(id%5) * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Wait for all pending rebalances
	time.Sleep(6 * time.Second)

	duration := time.Since(startTime)
	calls := repo.GetRebalanceCalls()

	t.Logf("Stress test completed in %v", duration)
	t.Logf("Total triggers: %d", totalTriggers)
	t.Logf("Actual rebalance calls: %d", calls)
	t.Logf("Reduction ratio: %.2f%%", (1.0-float64(calls)/float64(totalTriggers))*100)

	// Debouncing should reduce calls significantly
	if calls > int32(totalTriggers/10) {
		t.Errorf("Expected significant reduction in calls, got %d from %d triggers", calls, totalTriggers)
	}

	if calls < 1 {
		t.Errorf("Expected at least 1 rebalance call, got %d", calls)
	}
}
