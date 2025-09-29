package sse

import (
	"encoding/json"
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/rendis/structsconv"
)

type IChecklistItemUpdatesMapper interface {
	Map(domain.ChecklistItemUpdatesEvent) EventEnvelope
}

type checklistItemUpdatesMapper struct{}

func (mapper *checklistItemUpdatesMapper) Map(event domain.ChecklistItemUpdatesEvent) EventEnvelope {
	payload, error := mapPayload(event)
	if error != nil {
		panic(error)
	}
	return EventEnvelope{
		Type: EventEnvelopeType(event.EventType),
		Payload: &EventEnvelope_Payload{
			union: payload,
		},
	}
}

func mapPayload(event domain.ChecklistItemUpdatesEvent) (json.RawMessage, error) {
	source := event.Payload
	switch event.EventType {
	case domain.EventTypeChecklistItemCreated, domain.EventTypeChecklistItemUpdated:
		casted, ok := source.(domain.ChecklistItem)
		if !ok {
			return nil, fmt.Errorf("invalid payload type")
		}
		var createdPayload ChecklistItemResponse
		structsconv.Map(&casted, &createdPayload)
		b, _ := json.Marshal(createdPayload)
		return json.RawMessage(b), nil
	case domain.EventTypeChecklistItemDeleted:
		var deletedPayload ChecklistItemDeletedEventPayload
		casted, ok := source.(domain.ChecklistItemDeletedEventPayload)
		if !ok {
			return nil, fmt.Errorf("invalid payload type")
		}
		structsconv.Map(&casted, &deletedPayload)
		b, _ := json.Marshal(deletedPayload)
		return json.RawMessage(b), nil
	case domain.EventTypeChecklistItemRowDeleted:
		var rowDeletedPayload ChecklistItemRowDeletedEventPayload
		casted, ok := source.(domain.ChecklistItemRowDeletedPayload)
		if !ok {
			return nil, fmt.Errorf("invalid payload type")
		}
		structsconv.Map(&casted, &rowDeletedPayload)
		b, _ := json.Marshal(rowDeletedPayload)
		return json.RawMessage(b), nil
	case domain.EventTypeChecklistItemRowAdded:
		var rowAddedPayload ChecklistItemRowAddedEventPayload
		casted, ok := source.(domain.ChecklistItemRowAddedPayload)
		if !ok {
			return nil, fmt.Errorf("invalid payload type")
		}
		structsconv.Map(&casted, &rowAddedPayload)
		b, _ := json.Marshal(rowAddedPayload)
		return json.RawMessage(b), nil
	case domain.EventTypeChecklistItemReordered:
		var reorderedPayload ChecklistItemReorderedEventPayload
		casted, ok := source.(domain.ChecklistItemReorderedEventPayload)
		if !ok {
			return nil, fmt.Errorf("invalid payload type")
		}
		structsconv.Map(&casted, &reorderedPayload)
		b, _ := json.Marshal(reorderedPayload)
		return json.RawMessage(b), nil
	default:
		return nil, fmt.Errorf("unknown event type")
	}
}

func NewChecklistItemUpdatesMapper() IChecklistItemUpdatesMapper {
	return &checklistItemUpdatesMapper{}
}
