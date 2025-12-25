package domain

import "time"

type ChecklistInvite struct {
	Id          uint
	ChecklistId uint
	Name        *string // Optional friendly name for the invite
	InviteToken string
	CreatedBy   string // Google ID of the user who created the invite
	CreatedAt   time.Time
	ExpiresAt   *time.Time // nil means never expires
	ClaimedBy   *string    // Google ID of the user who claimed (nil if not claimed)
	ClaimedAt   *time.Time // nil if not claimed
	IsSingleUse bool       // If true, can only be claimed once
}
