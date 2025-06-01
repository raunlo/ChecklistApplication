package domain

type ChecklistItem struct {
	Id          uint
	Name        string
	Completed   bool
	Rows        []ChecklistItemRow
	OrderNumber uint
}

type ChecklistItemRow struct {
	Id        uint
	Name      string
	Completed bool
}
