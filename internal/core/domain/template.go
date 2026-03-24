package domain

import "time"

type Template struct {
	ID          uint
	UserID      string
	Name        string
	Description *string
	Rows        []TemplateRow
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TemplateRow struct {
	ID         uint
	TemplateID uint
	Name       string
	Position   float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
