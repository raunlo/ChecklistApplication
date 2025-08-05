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
