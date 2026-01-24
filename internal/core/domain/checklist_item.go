package domain

type ChecklistItem struct {
	Id          uint
	Name        string
	Completed   bool
	Rows        []ChecklistItemRow
	OrderNumber uint
	Position    float64
}

// Gap algorithm constants
const (
	DefaultGapSize    = 1000.0
	MinGapThreshold   = 0.001
	FirstItemPosition = 1000.0
)

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
