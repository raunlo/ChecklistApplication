package domain

import "time"

type Template struct {
	ID          uint
	UserID      string
	Name        string
	Description *string
	Items       []TemplateItem
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TemplateItem struct {
	ID        uint
	TemplateID uint
	Name      string
	Position  float64
	CreatedAt time.Time
	UpdatedAt time.Time
}
