package dbo

import "com.raunlo.checklist/internal/core/domain"

type ChecklistDbo struct {
	Id   uint   `primaryKey:"id"`
	Name string `db:"name"`
}

func MapChecklistDboToDomain(checklistDbo ChecklistDbo) domain.Checklist {
	return domain.Checklist{
		Id:   checklistDbo.Id,
		Name: checklistDbo.Name,
	}
}
