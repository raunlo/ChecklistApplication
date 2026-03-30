package repository

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
)

type ITemplateInviteRepository interface {
	CreateInvite(ctx context.Context, invite domain.TemplateInvite) (domain.TemplateInvite, domain.Error)
	FindInviteByToken(ctx context.Context, token string) (*domain.TemplateInvite, domain.Error)
	FindActiveInvitesByTemplateId(ctx context.Context, templateId uint) ([]domain.TemplateInvite, domain.Error)
	DeleteInviteById(ctx context.Context, inviteId uint) domain.Error
	ClaimInvite(ctx context.Context, token string, userId string) domain.Error
	ClaimInviteAndCreateShare(ctx context.Context, token string, userId string, templateId uint, sharedBy string) domain.Error
	DeleteExpiredInvites(ctx context.Context) (int64, domain.Error)
}
