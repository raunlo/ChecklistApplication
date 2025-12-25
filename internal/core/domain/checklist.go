package domain

type Checklist struct {
	Id             uint
	Name           string
	Owner          string
	ChecklistItems []ChecklistItem
	SharedWith     []string // List of user IDs this checklist is shared with
}
