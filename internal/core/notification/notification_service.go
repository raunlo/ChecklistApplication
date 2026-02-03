package notification

import (
	"context"
	"errors"
	"log"
	"sync"

	"com.raunlo.checklist/internal/core/domain"
	guardrail "com.raunlo.checklist/internal/core/guard_rail"
)

type INotificationService interface {
	NotifyItemCreated(ctx context.Context, checklistId uint, item domain.ChecklistItem)
	NotifyItemUpdated(ctx context.Context, checklistId uint, item domain.ChecklistItem)
	NotifyItemDeleted(ctx context.Context, checklistId uint, itemId uint)
	NotifyItemSoftDeleted(ctx context.Context, checklistId uint, itemId uint)
	NotifyItemRestored(ctx context.Context, checklistId uint, item domain.ChecklistItem)
	NotifyItemRowAdded(ctx context.Context, checklistId uint, itemId uint, row domain.ChecklistItemRow)
	NotifyItemRowDeleted(ctx context.Context, checklistId uint, itemId uint, rowId uint)
	NotifyItemReordered(ctx context.Context, request domain.ChangeOrderRequest, resp domain.ChangeOrderResponse)
}

type notificationService struct {
	broker IBroker
}

func NewNotificationService(notificationBroker IBroker) INotificationService {
	return &notificationService{
		broker: notificationBroker,
	}
}

func (n *notificationService) NotifyItemCreated(ctx context.Context, checklistId uint, item domain.ChecklistItem) {
	n.broker.Publish(ctx, checklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemCreated,
		Payload:   item,
	})
}

func (n *notificationService) NotifyItemUpdated(ctx context.Context, checklistId uint, item domain.ChecklistItem) {
	n.broker.Publish(ctx, checklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemUpdated,
		Payload:   item,
	})
}

func (n *notificationService) NotifyItemDeleted(ctx context.Context, checklistId uint, itemId uint) {
	n.broker.Publish(ctx, checklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemDeleted,
		Payload:   domain.ChecklistItemDeletedEventPayload{ItemId: itemId},
	})
}

func (n *notificationService) NotifyItemSoftDeleted(ctx context.Context, checklistId uint, itemId uint) {
	n.broker.Publish(ctx, checklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemSoftDeleted,
		Payload:   domain.ChecklistItemSoftDeletedEventPayload{ItemId: itemId},
	})
}

func (n *notificationService) NotifyItemRestored(ctx context.Context, checklistId uint, item domain.ChecklistItem) {
	n.broker.Publish(ctx, checklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemRestored,
		Payload:   domain.ChecklistItemRestoredEventPayload{Item: item},
	})
}

func (n *notificationService) NotifyItemRowAdded(ctx context.Context, checklistId uint, itemId uint, row domain.ChecklistItemRow) {
	n.broker.Publish(ctx, checklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemRowAdded,
		Payload: domain.ChecklistItemRowAddedPayload{
			ItemId: itemId,
			Row:    row,
		},
	})
}

func (n *notificationService) NotifyItemRowDeleted(ctx context.Context, checklistId uint, itemId uint, rowId uint) {
	n.broker.Publish(ctx, checklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemRowDeleted,
		Payload: domain.ChecklistItemRowDeletedPayload{
			RowId:  rowId,
			ItemId: itemId,
		},
	})
}

func (n *notificationService) NotifyItemReordered(ctx context.Context, request domain.ChangeOrderRequest, resp domain.ChangeOrderResponse) {
	n.broker.Publish(ctx, request.ChecklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemReordered,
		Payload: domain.ChecklistItemReorderedEventPayload{
			ItemId:         request.ChecklistItemId,
			NewOrderNumber: resp.OrderNumber,
			OrderChanged:   true,
		},
	})
}

type IBroker interface {
	// Subscribe registers a new client and returns a channel to receive messages.
	Subscribe(ctx context.Context, checklistId uint) (chan domain.ChecklistItemUpdatesEvent, error)
	// Unsubscribe removes a client channel.
	Unsubscribe(ctx context.Context, checklistId uint) error
	// Publish sends a message to all subscribed clients for a checklistId. Non-blocking runs in a goroutine.
	Publish(ctx context.Context, checklistId uint, event domain.ChecklistItemUpdatesEvent)
}

// clientChannel wraps a channel with close-once semantics to prevent double-close panics
type clientChannel struct {
	ch        chan domain.ChecklistItemUpdatesEvent
	closeOnce sync.Once
	closed    bool
}

func (cc *clientChannel) Close() {
	cc.closeOnce.Do(func() {
		cc.closed = true
		close(cc.ch)
	})
}

func (cc *clientChannel) Send(event domain.ChecklistItemUpdatesEvent) bool {
	if cc.closed {
		return false
	}
	select {
	case cc.ch <- event:
		return true
	default:
		return false
	}
}

type broker struct {
	// map of checklistIds to *sync.Map of client channels
	clients            sync.Map // key: uint -> value: *sync.Map (key: clientId string -> value: *clientChannel)
	checklistGuardrail guardrail.IChecklistOwnershipChecker
}

func NewBroker(guardrail guardrail.IChecklistOwnershipChecker) IBroker {
	return &broker{
		checklistGuardrail: guardrail,
	}
}

// Subscribe registers a client and returns a channel to receive messages
func (b *broker) Subscribe(ctx context.Context, checklistId uint) (chan domain.ChecklistItemUpdatesEvent, error) {
	if err := b.checklistGuardrail.HasAccessToChecklist(ctx, checklistId); err != nil {
		return nil, err
	}
	clientId := ctx.Value(domain.ClientIdContextKey)
	if clientId == nil {
		return nil, errors.New("ClientID not found")
	}

	// Close any existing channel for this client before creating a new one
	newInner := &sync.Map{}
	actual, _ := b.clients.LoadOrStore(checklistId, newInner)
	inner := actual.(*sync.Map)

	// If there's an existing channel for this client, close it first
	if existing, loaded := inner.Load(clientId); loaded {
		if cc, ok := existing.(*clientChannel); ok {
			cc.Close()
		}
	}

	cc := &clientChannel{
		ch: make(chan domain.ChecklistItemUpdatesEvent, 10),
	}
	inner.Store(clientId, cc)
	return cc.ch, nil
}

// Unsubscribe removes a client channel
func (b *broker) Unsubscribe(ctx context.Context, checklistId uint) error {
	clientId := ctx.Value(domain.ClientIdContextKey)
	if clientId == nil {
		return errors.New("clientId not found in context")
	}
	val, ok := b.clients.Load(checklistId)
	if !ok {
		return errors.New("no clients found for checklistId")
	}
	inner := val.(*sync.Map)
	// Remove the channel from the inner map and close it safely
	if existing, loaded := inner.LoadAndDelete(clientId); loaded {
		if cc, ok := existing.(*clientChannel); ok {
			cc.Close()
		}
	}
	return nil
}

// Publish sends message to the broker (non-blocking). If out buffer is full, event is dropped.
func (b *broker) Publish(ctx context.Context, checklistId uint, event domain.ChecklistItemUpdatesEvent) {
	go func() {
		clientIdFromContext := ctx.Value(domain.ClientIdContextKey)
		if clientIdFromContext == nil {
			log.Print("sse: no clientId in context, skipping publish")
			return
		}
		val, ok := b.clients.Load(checklistId)
		if !ok {
			return
		}
		inner := val.(*sync.Map)
		inner.Range(func(clientId any, v any) bool {
			cc, ok := v.(*clientChannel)
			if !ok {
				return true
			}
			// Skip sending to the originating client
			if clientIdFromContext == clientId.(string) {
				return true
			}
			// Use the safe Send method which handles closed channels
			if !cc.Send(event) {
				// Buffer is full or channel closed, try overflow notification
				log.Printf("sse: buffer full for client %v, sending overflow notification", clientId)
				overflowEvent := domain.ChecklistItemUpdatesEvent{
					EventType: domain.EventTypeBufferOverflow,
					Payload: domain.BufferOverflowEventPayload{
						Message: "Event buffer full, please refresh to ensure data consistency",
					},
				}
				if !cc.Send(overflowEvent) {
					log.Printf("sse: dropping overflow notification for client %v, client too slow or disconnected", clientId)
				}
			}
			return true
		})
	}()
}
