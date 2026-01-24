package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"com.raunlo.checklist/internal/core/repository"
)

// IRebalanceService handles async rebalancing of checklist item positions
type IRebalanceService interface {
	// TriggerRebalance schedules a rebalance for an entire checklist
	TriggerRebalance(checklistId uint)
}

// pendingRebalance tracks a pending rebalance with both debounce and max-wait timers
type pendingRebalance struct {
	debounceTimer *time.Timer
	maxWaitTimer  *time.Timer
	mu            sync.Mutex
	executed      bool
}

type rebalanceService struct {
	repository repository.IChecklistItemsRepository
	pending    sync.Map // key: checklistId -> *pendingRebalance
	debounceMs time.Duration
	maxWaitMs  time.Duration
}

// NewRebalanceService creates a new rebalance service
func NewRebalanceService(repo repository.IChecklistItemsRepository) IRebalanceService {
	return &rebalanceService{
		repository: repo,
		debounceMs: 500 * time.Millisecond,
		maxWaitMs:  5 * time.Second, // Guarantee execution within 5 seconds
	}
}

func (s *rebalanceService) TriggerRebalance(checklistId uint) {
	key := fmt.Sprintf("%d", checklistId)

	// Check if there's already a pending rebalance
	if val, ok := s.pending.Load(key); ok {
		pending := val.(*pendingRebalance)
		pending.mu.Lock()

		// If already executed, delete and start fresh
		if pending.executed {
			pending.mu.Unlock()
			s.pending.Delete(key)
			// Fall through to create new pending
		} else {
			// Reset debounce timer but keep max-wait timer
			pending.debounceTimer.Stop()
			pending.debounceTimer = time.AfterFunc(s.debounceMs, func() {
				s.executeOnce(key, checklistId)
			})
			pending.mu.Unlock()
			return
		}
	}

	// Create new pending rebalance
	pending := &pendingRebalance{}

	executeFunc := func() {
		s.executeOnce(key, checklistId)
	}

	pending.debounceTimer = time.AfterFunc(s.debounceMs, executeFunc)
	pending.maxWaitTimer = time.AfterFunc(s.maxWaitMs, executeFunc)

	s.pending.Store(key, pending)
}

// executeOnce ensures rebalance runs exactly once, then cleans up
func (s *rebalanceService) executeOnce(key string, checklistId uint) {
	val, ok := s.pending.Load(key)
	if !ok {
		return
	}

	pending := val.(*pendingRebalance)
	pending.mu.Lock()

	// Check if already executed
	if pending.executed {
		pending.mu.Unlock()
		return
	}
	pending.executed = true

	// Stop both timers
	pending.debounceTimer.Stop()
	pending.maxWaitTimer.Stop()
	pending.mu.Unlock()

	// Execute rebalance
	s.executeRebalance(checklistId)

	// Cleanup
	s.pending.Delete(key)
}

func (s *rebalanceService) executeRebalance(checklistId uint) {
	ctx := context.Background()

	err := s.repository.RebalancePositions(ctx, checklistId)
	if err != nil {
		log.Printf("rebalance: failed for checklist %d: %v", checklistId, err)
	} else {
		log.Printf("rebalance: completed for checklist %d", checklistId)
	}
}
