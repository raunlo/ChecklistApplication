package domain

import "time"

type TemplateInvite struct {
	Id          uint
	TemplateId  uint
	Name        *string
	InviteToken string
	CreatedBy   string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	ClaimedBy   *string
	ClaimedAt   *time.Time
	IsSingleUse bool
}
