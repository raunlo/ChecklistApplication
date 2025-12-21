package repository

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
)

type IChecklistRepository interface {
	UpdateChecklist(ctx context.Context, checklist domain.Checklist) (domain.Checklist, domain.Error)
	SaveChecklist(ctx context.Context, checklist domain.Checklist) (domain.Checklist, domain.Error)
	FindChecklistById(ctx context.Context, id uint) (*domain.Checklist, domain.Error)
	DeleteChecklistById(ctx context.Context, id uint) domain.Error
	CheckUserHasAccessToChecklist(ctx context.Context, checklistId uint, userId string) (bool, domain.Error)
	CheckUserIsOwner(ctx context.Context, checklistId uint, userId string) (bool, domain.Error)
	FindAllChecklists(ctx context.Context) ([]domain.Checklist, domain.Error)
	CreateChecklistShare(ctx context.Context, checklistId uint, sharedByUserId string, sharedWithUserId string) domain.Error
}
