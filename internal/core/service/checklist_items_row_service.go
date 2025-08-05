package service

import "com.raunlo.checklist/internal/core/domain"

type IChecklistItemRowsService interface {
	AddNewRow(request domain.CreateChecklistItemRow) domain.ChecklistItemRow
}

type checklistItemRowsService struct {
}

func (s *checklistItemRowsService) AddNewRow(request domain.CreateChecklistItemRow) domain.ChecklistItemRow {

}
