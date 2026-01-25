package domain

// User represents a user in the system
// Audit fields (created_at, updated_at) are managed by SQL DEFAULT CURRENT_TIMESTAMP
type User struct {
	UserId string
	Name   string
}
