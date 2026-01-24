package dbo

import "com.raunlo.checklist/internal/core/domain"

type ChecklistItemDbo struct {
	Id          uint                  `primaryKey:"checklist_item_id"`
	Name        string                `db:"checklist_item_name"`
	Completed   bool                  `db:"checklist_item_completed"`
	Rows        []ChecklistItemRowDbo `relationship:"oneToMany"`
	OrderNumber uint                  `db:"order_number"`
	Position    float64               `db:"position"`
}

type ChecklistItemRowDbo struct {
	Id        uint   `primaryKey:"checklist_item_row_id"`
	Name      string `db:"checklist_item_row_name"`
	Completed bool   `db:"checklist_item_row_completed"`
}

func MapChecklistItemDboToDomain(checklistItemDbo ChecklistItemDbo) domain.ChecklistItem {
	var checklistItemRows []domain.ChecklistItemRow
	for _, row := range checklistItemDbo.Rows {
		checklistItemRows = append(checklistItemRows, MapChecklistItemRowsDboToDomain(row))
	}

	return domain.ChecklistItem{
		Id:          checklistItemDbo.Id,
		Name:        checklistItemDbo.Name,
		Completed:   checklistItemDbo.Completed,
		Rows:        checklistItemRows,
		OrderNumber: checklistItemDbo.OrderNumber,
		Position:    checklistItemDbo.Position,
	}
}

func MapChecklistItemRowsDboToDomain(checklistItemRowDbo ChecklistItemRowDbo) domain.ChecklistItemRow {
	return domain.ChecklistItemRow{
		Id:        checklistItemRowDbo.Id,
		Name:      checklistItemRowDbo.Name,
		Completed: checklistItemRowDbo.Completed,
	}
}
