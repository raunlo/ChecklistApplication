package repository

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
)

type ITemplateRepository interface {
	SaveTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error)
	FindTemplateById(ctx context.Context, id uint) (*domain.Template, domain.Error)
	FindTemplatesByUserId(ctx context.Context, userId string) ([]domain.Template, domain.Error)
	UpdateTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error)
	DeleteTemplate(ctx context.Context, id uint) domain.Error
	CheckUserIsTemplateOwner(ctx context.Context, templateId uint, userId string) (bool, domain.Error)
}
