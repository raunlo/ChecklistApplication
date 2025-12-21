package repository

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
)

type IChecklistInviteRepository interface {
	CreateInvite(ctx context.Context, invite domain.ChecklistInvite) (domain.ChecklistInvite, domain.Error)
	FindInviteByToken(ctx context.Context, token string) (*domain.ChecklistInvite, domain.Error)
	FindActiveInvitesByChecklistId(ctx context.Context, checklistId uint) ([]domain.ChecklistInvite, domain.Error)
	DeleteInviteById(ctx context.Context, inviteId uint) domain.Error
	ClaimInvite(ctx context.Context, token string, userId string) domain.Error
	DeleteExpiredInvites(ctx context.Context) (int64, domain.Error)
}
