package repository

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
)

type IChecklistRepository interface {
	UpdateChecklist(checklist domain.Checklist) (domain.Checklist, domain.Error)
	SaveChecklist(checklist domain.Checklist) (domain.Checklist, domain.Error)
	FindChecklistById(id uint) (*domain.Checklist, domain.Error)
	DeleteChecklistById(id uint) domain.Error
	CheckUserHasAccessToChecklist(checklistId uint, userId string) (bool, domain.Error)
	FindAllChecklists(ctx context.Context) ([]domain.Checklist, domain.Error)
}
