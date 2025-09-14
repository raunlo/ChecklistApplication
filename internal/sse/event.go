package sse

// Event is the typed payload sent over SSE. Keeping Payload as interface{}
// for flexibility; prefer structured payloads where possible.
type Event struct {
	Type        string      `json:"type"`
	ChecklistID int         `json:"checklistId"`
	Payload     interface{} `json:"payload,omitempty"`
	ClientID    string      `json:"clientId,omitempty"`
	ID          *int        `json:"id,omitempty"`
}
