package domain

type ChecklistItem struct {
	Id          uint
	Name        string
	Completed   bool
	Rows        []ChecklistItemRow
	OrderNumber uint
}

const PhantomChecklsitItemName = "__phantom__"

type ChecklistItemRow struct {
	Id        uint
	Name      string
	Completed bool
}

// ChecklistItemRowDeletionResult contains information about a row deletion operation
type ChecklistItemRowDeletionResult struct {
	Success           bool // Whether the deletion was successful
	ItemAutoCompleted bool // Whether the parent item was automatically marked as completed
}
