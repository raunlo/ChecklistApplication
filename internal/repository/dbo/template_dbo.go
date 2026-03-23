package dbo

import (
	"time"

	"com.raunlo.checklist/internal/core/domain"
)

type TemplateDBO struct {
	ID          uint64     `db:"ID"`
	UserID      string     `db:"USER_ID"`
	Name        string     `db:"NAME"`
	Description *string    `db:"DESCRIPTION"`
	CreatedAt   time.Time  `db:"CREATED_AT"`
	UpdatedAt   time.Time  `db:"UPDATED_AT"`
}

type TemplateItemDBO struct {
	ID        uint64    `db:"ID"`
	TemplateID uint64    `db:"TEMPLATE_ID"`
	Name      string    `db:"NAME"`
	Position  float64   `db:"POSITION"`
	CreatedAt time.Time `db:"CREATED_AT"`
	UpdatedAt time.Time `db:"UPDATED_AT"`
}

func (t *TemplateDBO) ToDomain() domain.Template {
	return domain.Template{
		ID:          uint(t.ID),
		UserID:      t.UserID,
		Name:        t.Name,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		Items:       []domain.TemplateItem{},
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

func (ti *TemplateItemDBO) ToDomain() domain.TemplateItem {
	return domain.TemplateItem{
		ID:        uint(ti.ID),
		TemplateID: uint(ti.TemplateID),
		Name:      ti.Name,
		Position:  ti.Position,
		CreatedAt: ti.CreatedAt,
		UpdatedAt: ti.UpdatedAt,
	}
}

func (ti *TemplateItemDBO) FromDomain(item domain.TemplateItem) {
	ti.ID = uint64(item.ID)
	ti.TemplateID = uint64(item.TemplateID)
	ti.Name = item.Name
	ti.Position = item.Position
	ti.CreatedAt = item.CreatedAt
	ti.UpdatedAt = item.UpdatedAt
}
