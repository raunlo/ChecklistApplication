package dbo

import (
	"time"

	"com.raunlo.checklist/internal/core/domain"
)

type ChecklistInviteDbo struct {
	Id          uint       `primaryKey:"id"`
	ChecklistId uint       `db:"checklist_id"`
	InviteToken string     `db:"invite_token"`
	CreatedBy   string     `db:"created_by"`
	CreatedAt   time.Time  `db:"created_at"`
	ExpiresAt   *time.Time `db:"expires_at"`
	ClaimedBy   *string    `db:"claimed_by"`
	ClaimedAt   *time.Time `db:"claimed_at"`
	IsSingleUse bool       `db:"is_single_use"`
}

func MapChecklistInviteDboToDomain(dbo ChecklistInviteDbo) domain.ChecklistInvite {
	return domain.ChecklistInvite{
		Id:          dbo.Id,
		ChecklistId: dbo.ChecklistId,
		InviteToken: dbo.InviteToken,
		CreatedBy:   dbo.CreatedBy,
		CreatedAt:   dbo.CreatedAt,
		ExpiresAt:   dbo.ExpiresAt,
		ClaimedBy:   dbo.ClaimedBy,
		ClaimedAt:   dbo.ClaimedAt,
		IsSingleUse: dbo.IsSingleUse,
	}
}

func MapDomainToChecklistInviteDbo(invite domain.ChecklistInvite) ChecklistInviteDbo {
	return ChecklistInviteDbo{
		Id:          invite.Id,
		ChecklistId: invite.ChecklistId,
		InviteToken: invite.InviteToken,
		CreatedBy:   invite.CreatedBy,
		CreatedAt:   invite.CreatedAt,
		ExpiresAt:   invite.ExpiresAt,
		ClaimedBy:   invite.ClaimedBy,
		ClaimedAt:   invite.ClaimedAt,
		IsSingleUse: invite.IsSingleUse,
	}
}
