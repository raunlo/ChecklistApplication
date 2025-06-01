package repository

import "com.raunlo.checklist/internal/core/domain"

type IChecklistRepository interface {
	UpdateChecklist(checklist domain.Checklist) (domain.Checklist, domain.Error)
	SaveChecklist(checklist domain.Checklist) (domain.Checklist, domain.Error)
	FindChecklistById(id uint) (*domain.Checklist, domain.Error)
	DeleteChecklistById(id uint) domain.Error
	FindAllChecklists() ([]domain.Checklist, domain.Error)
}
