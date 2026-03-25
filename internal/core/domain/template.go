package domain

import "time"

type Template struct {
	Id          uint
	UserId      string
	Name        string
	Description *string
	Rows        []TemplateRow
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TemplateRow struct {
	Id         uint
	TemplateId uint
	Name       string
	Position   float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
