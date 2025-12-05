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

type broker struct {
	// map of checklistIds to *sync.Map of client channels
	clients            sync.Map // key: uint -> value: *sync.Map (key: chan domain.ChecklistItemUpdatesEvent -> value: struct{})
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
	ch := make(chan domain.ChecklistItemUpdatesEvent, 10)
	newInner := &sync.Map{}
	actual, _ := b.clients.LoadOrStore(checklistId, newInner)
	inner := actual.(*sync.Map)
	inner.Store(clientId, ch)
	return ch, nil
}

// Unsubscribe removes a client channel
func (b *broker) Unsubscribe(ctx context.Context, checklistId uint) error {
	clientId := ctx.Value(domain.ClientIdContextKey)
	if clientId != nil {
		return errors.New("clientId found in context: " + clientId.(string))
	}
	val, ok := b.clients.Load(checklistId)
	if !ok {
		return errors.New("no clients found for checklistId")
	}
	var closeOnce sync.Once
	inner := val.(*sync.Map)
	// Remove the channel from the inner map
	ch, _ := inner.LoadAndDelete(clientId)
	// Close the channel safely
	closeOnce.Do(func() {
		if ch != nil {
			close(ch.(chan domain.ChecklistItemUpdatesEvent))
		}
	})
	return nil
}

// Publish sends message to the broker (non-blocking). If out buffer is full, event is dropped.
func (b *broker) Publish(ctx context.Context, checklistId uint, event domain.ChecklistItemUpdatesEvent) {
	go func() {
		clientIdFromContext := ctx.Value(domain.ClientIdContextKey)
		if clientIdFromContext == nil {
			log.Print("sse: no clientId in context, will publish to all clients")
			return
		}
		val, ok := b.clients.Load(checklistId)
		if !ok {
			return
		}
		inner := val.(*sync.Map)
		inner.Range(func(clientId any, v any) bool {
			ch, ok := v.(chan domain.ChecklistItemUpdatesEvent)
			if !ok {
				return true
			}
			if clientIdFromContext != nil && clientIdFromContext == clientId.(string) {
				return true
			}
			// send in a safe way to avoid panic if channel is closed concurrently
			func() {
				defer func() {
					if r := recover(); r != nil {
						// someone closed the channel concurrently; ignore
					}
				}()
				select {
				case ch <- event:
				default:
					log.Printf("sse: dropping event, client buffer full")
				}
			}()
			return true
		})
	}()
}
