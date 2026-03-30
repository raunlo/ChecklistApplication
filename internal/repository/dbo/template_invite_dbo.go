package dbo

import (
	"time"

	"com.raunlo.checklist/internal/core/domain"
)

type TemplateInviteDbo struct {
	Id          uint       `primaryKey:"id"`
	TemplateId  uint       `db:"template_id"`
	Name        *string    `db:"name"`
	InviteToken string     `db:"invite_token"`
	CreatedBy   string     `db:"created_by"`
	CreatedAt   time.Time  `db:"created_at"`
	ExpiresAt   *time.Time `db:"expires_at"`
	ClaimedBy   *string    `db:"claimed_by"`
	ClaimedAt   *time.Time `db:"claimed_at"`
	IsSingleUse bool       `db:"is_single_use"`
}

func MapTemplateInviteDboToDomain(dbo TemplateInviteDbo) domain.TemplateInvite {
	return domain.TemplateInvite{
		Id:          dbo.Id,
		TemplateId:  dbo.TemplateId,
		Name:        dbo.Name,
		InviteToken: dbo.InviteToken,
		CreatedBy:   dbo.CreatedBy,
		CreatedAt:   dbo.CreatedAt,
		ExpiresAt:   dbo.ExpiresAt,
		ClaimedBy:   dbo.ClaimedBy,
		ClaimedAt:   dbo.ClaimedAt,
		IsSingleUse: dbo.IsSingleUse,
	}
}
