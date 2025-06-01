package domain

type Checklist struct {
	Id             uint
	Name           string
	ChecklistItems []ChecklistItem
}
