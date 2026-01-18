package domain

const (
	EventTypeChecklistItemCreated    = "checklistItemCreated"
	EventTypeChecklistItemUpdated    = "checklistItemUpdated"
	EventTypeChecklistItemToggled    = "checklistItemToggled"
	EventTypeChecklistItemReordered  = "checklistItemReordered"
	EventTypeChecklistItemDeleted    = "checklistItemDeleted"
	EventTypeChecklistItemRowDeleted = "checklistItemRowDeleted"
	EventTypeChecklistItemRowAdded   = "checklistItemRowAdded"
	EventTypeBufferOverflow          = "bufferOverflow"
)

type ChecklistItemToggledEventPayload struct {
	ItemId    uint `json:"itemId"`
	Completed bool `json:"completed"`
}

type ChecklistItemReorderedEventPayload struct {
	ItemId         uint `json:"itemId"`
	NewOrderNumber uint `json:"newOrderNumber"`
	OrderChanged   bool `json:"orderChanged"`
}

type ChecklistItemDeletedEventPayload struct {
	ItemId uint `json:"itemId"`
}

type ChecklistItemRowAddedPayload struct {
	ItemId uint             `json:"itemId"`
	Row    ChecklistItemRow `json:"row"`
}

type ChecklistItemRowDeletedPayload struct {
	RowId  uint `json:"rowId"`
	ItemId uint `json:"itemId"`
}

type BufferOverflowEventPayload struct {
	Message string `json:"message"`
}

type ChecklistItemUpdatesEvent struct {
	EventType string
	Payload   interface{}
}
