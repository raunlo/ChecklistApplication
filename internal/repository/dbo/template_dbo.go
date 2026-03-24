package dbo

import (
	"time"

	"com.raunlo.checklist/internal/core/domain"
)

type TemplateDBO struct {
	ID          uint64    `db:"ID"`
	UserID      string    `db:"USER_ID"`
	Name        string    `db:"NAME"`
	Description *string   `db:"DESCRIPTION"`
	CreatedAt   time.Time `db:"CREATED_AT"`
	UpdatedAt   time.Time `db:"UPDATED_AT"`
}

type TemplateRowDBO struct {
	ID         uint64    `db:"ID"`
	TemplateID uint64    `db:"TEMPLATE_ID"`
	Name       string    `db:"NAME"`
	Position   float64   `db:"POSITION"`
	CreatedAt  time.Time `db:"CREATED_AT"`
	UpdatedAt  time.Time `db:"UPDATED_AT"`
}

func (t *TemplateDBO) ToDomain() domain.Template {
	return domain.Template{
		ID:          uint(t.ID),
		UserID:      t.UserID,
		Name:        t.Name,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		Rows:        []domain.TemplateRow{},
	}
}

func (t *TemplateDBO) FromDomain(template domain.Template) {
	t.ID = uint64(template.ID)
	t.UserID = template.UserID
	t.Name = template.Name
	t.Description = template.Description
	t.CreatedAt = template.CreatedAt
	t.UpdatedAt = template.UpdatedAt
}

func (r *TemplateRowDBO) ToDomain() domain.TemplateRow {
	return domain.TemplateRow{
		ID:         uint(r.ID),
		TemplateID: uint(r.TemplateID),
		Name:       r.Name,
		Position:   r.Position,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

func (r *TemplateRowDBO) FromDomain(row domain.TemplateRow) {
	r.ID = uint64(row.ID)
	r.TemplateID = uint64(row.TemplateID)
	r.Name = row.Name
	r.Position = row.Position
	r.CreatedAt = row.CreatedAt
	r.UpdatedAt = row.UpdatedAt
}
