package guardrail

import (
	"context"
	"log"

	"com.raunlo.checklist/internal/core/domain"
	domainErr "com.raunlo.checklist/internal/core/error"
	"com.raunlo.checklist/internal/core/repository"
)

type ITemplateOwnershipChecker interface {
	IsTemplateOwner(ctx context.Context, templateId uint) domain.Error
}

type templateOwnershipCheckerService struct {
	repository repository.ITemplateRepository
}

func (service *templateOwnershipCheckerService) IsTemplateOwner(ctx context.Context, templateId uint) domain.Error {
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return err
	}

	isOwner, err := service.repository.CheckUserIsTemplateOwner(ctx, templateId, userId)
	log.Printf("GuardRail: User(id=%s) owner check for template %d: %v", domain.GetHashedUserIdFromContext(ctx), templateId, isOwner)
	if err != nil {
		return domain.Wrap(err, "Failed to check template ownership", 500)
	}
	if !isOwner {
		return domainErr.NewTemplateNotFoundError(templateId)
	}

	return nil
}

func NewTemplateOwnershipCheckerService(repository repository.ITemplateRepository) ITemplateOwnershipChecker {
	return &templateOwnershipCheckerService{
		repository: repository,
	}
}
