package notification

import (
	"log"
	"sync"

	"com.raunlo.checklist/internal/core/domain"
)

type INotificationService interface {
	NotifyItemCreated(checklistId uint, item domain.ChecklistItem)
	NotifyItemUpdated(checklistId uint, item domain.ChecklistItem)
	NotifyItemDeleted(checklistId uint, itemId uint)
	NotifyItemRowAdded(checklistId uint, itemId uint, row domain.ChecklistItemRow)
	NotifyItemRowDeleted(checklistId uint, itemId uint, rowId uint)
	NotifyItemReordered(request domain.ChangeOrderRequest, resp domain.ChangeOrderResponse)
}

type notificationService struct {
	broker IBroker
}

func NewNotificationService(notificationBroker IBroker) INotificationService {
	return &notificationService{
		broker: notificationBroker,
	}
}

func (n *notificationService) NotifyItemCreated(checklistId uint, item domain.ChecklistItem) {
	n.broker.Publish(checklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemCreated,
		Payload:   item,
	})
}

func (n *notificationService) NotifyItemUpdated(checklistId uint, item domain.ChecklistItem) {
	n.broker.Publish(checklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemUpdated,
		Payload:   item,
	})
}

func (n *notificationService) NotifyItemDeleted(checklistId uint, itemId uint) {
	n.broker.Publish(checklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemDeleted,
		Payload:   domain.ChecklistItemDeletedEventPayload{ItemId: itemId},
	})
}

func (n *notificationService) NotifyItemRowAdded(checklistId uint, itemId uint, row domain.ChecklistItemRow) {
	n.broker.Publish(checklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemRowAdded,
		Payload: domain.ChecklistItemRowAddedPayload{
			ItemId: itemId,
			Row:    row,
		},
	})
}

func (n *notificationService) NotifyItemRowDeleted(checklistId uint, itemId uint, rowId uint) {
	n.broker.Publish(checklistId, domain.ChecklistItemUpdatesEvent{
		EventType: domain.EventTypeChecklistItemRowDeleted,
		Payload: domain.ChecklistItemRowDeletedPayload{
			RowId:  rowId,
			ItemId: itemId,
		},
	})
}

func (n *notificationService) NotifyItemReordered(request domain.ChangeOrderRequest, resp domain.ChangeOrderResponse) {
	n.broker.Publish(request.ChecklistId, domain.ChecklistItemUpdatesEvent{
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
	Subscribe(checklistId uint) chan domain.ChecklistItemUpdatesEvent
	// Unsubscribe removes a client channel.
	Unsubscribe(checklistId uint, ch chan domain.ChecklistItemUpdatesEvent)
	// Publish sends a message to all subscribed clients for a checklistId. Non-blocking runs in a goroutine.
	Publish(checklistId uint, event domain.ChecklistItemUpdatesEvent)
}

type broker struct {
	// map of checklistIds to map of client channels
	clients map[uint]map[chan domain.ChecklistItemUpdatesEvent]struct{}
	lock    sync.Mutex
}

func NewBroker() IBroker {
	return &broker{
		clients: make(map[uint]map[chan domain.ChecklistItemUpdatesEvent]struct{}),
	}
}

// Subscribe registers a client and returns a channel to receive messages
func (b *broker) Subscribe(checklistId uint) chan domain.ChecklistItemUpdatesEvent {
	b.lock.Lock()
	defer b.lock.Unlock()
	ch := make(chan domain.ChecklistItemUpdatesEvent, 10)
	if _, ok := b.clients[checklistId]; !ok {
		b.clients[checklistId] = make(map[chan domain.ChecklistItemUpdatesEvent]struct{})
	}
	b.clients[checklistId][ch] = struct{}{}
	return ch
}

// Unsubscribe removes a client channel
func (b *broker) Unsubscribe(checklistId uint, ch chan domain.ChecklistItemUpdatesEvent) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if clients, ok := b.clients[checklistId]; ok {
		if _, exists := clients[ch]; exists {
			delete(clients, ch)
			close(ch)
		}
	}
}

// Publish sends message to the broker (non-blocking). If out buffer is full, event is dropped.
func (b *broker) Publish(checklistId uint, event domain.ChecklistItemUpdatesEvent) {
	// Marshal event to JSON before starting the goroutine

	go func() {
		b.lock.Lock()
		channels, ok := b.clients[checklistId]
		b.lock.Unlock()
		if !ok {
			return
		}
		for ch := range channels {
			select {
			case ch <- event:
			default:
				log.Printf("sse: dropping event, client buffer full")
			}
		}
	}()
}
