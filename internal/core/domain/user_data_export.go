package domain

import "time"

type UserDataExport struct {
	UserId     string
	ExportedAt time.Time
	Checklists []ExportedChecklist
}

type ExportedChecklist struct {
	Id        uint
	Name      string
	CreatedAt time.Time
	Items     []ExportedChecklistItem
	Shares    []ExportedChecklistShare
}

type ExportedChecklistItem struct {
	Id          uint
	Name        string
	Completed   bool
	OrderNumber int
	Rows        []ExportedChecklistItemRow
}

type ExportedChecklistItemRow struct {
	Id        uint
	Name      string
	Completed bool
}

type ExportedChecklistShare struct {
	SharedWithUserId string
	PermissionLevel  string
	SharedAt         time.Time
}
