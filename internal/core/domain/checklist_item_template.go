package domain

type ChecklistItemTemplate struct {
	Id   uint
	Rows []ChecklistItemTemplateRow
}

type ChecklistItemTemplateRow struct {
	Id uint
}
