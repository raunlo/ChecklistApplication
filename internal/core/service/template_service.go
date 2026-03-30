package service

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	coreError "com.raunlo.checklist/internal/core/error"
	guardrail "com.raunlo.checklist/internal/core/guard_rail"
	"com.raunlo.checklist/internal/core/repository"
)

type ITemplateService interface {
	SaveTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error)
	FindTemplateById(ctx context.Context, id uint) (*domain.Template, domain.Error)
	FindAllTemplates(ctx context.Context) ([]domain.Template, domain.Error)
	UpdateTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error)
	DeleteTemplate(ctx context.Context, id uint) domain.Error
	CreateTemplateFromItem(ctx context.Context, checklistId uint, name string, description *string, checklistItemId uint) (domain.Template, domain.Error)
	ApplyTemplateToChecklist(ctx context.Context, checklistId uint, templateId uint) (domain.ChecklistItem, domain.Error)
	LeaveSharedTemplate(ctx context.Context, templateId uint) domain.Error
}

type templateService struct {
	templateRepository        repository.ITemplateRepository
	templateOwnershipChecker  guardrail.ITemplateOwnershipChecker
	checklistItemService      IChecklistItemsService
	checklistOwnershipChecker guardrail.IChecklistOwnershipChecker
}

func (service *templateService) SaveTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error) {
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return domain.Template{}, err
	}

	template.UserId = userId
	return service.templateRepository.SaveTemplate(ctx, template)
}

func (service *templateService) FindTemplateById(ctx context.Context, id uint) (*domain.Template, domain.Error) {
	return service.templateRepository.FindTemplateById(ctx, id)
}

func (service *templateService) FindAllTemplates(ctx context.Context) ([]domain.Template, domain.Error) {
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return service.templateRepository.FindTemplatesByUserId(ctx, userId)
}

func (service *templateService) UpdateTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error) {
	return service.templateRepository.UpdateTemplate(ctx, template)
}

func (service *templateService) DeleteTemplate(ctx context.Context, id uint) domain.Error {
	return service.templateRepository.DeleteTemplate(ctx, id)
}

func (service *templateService) CreateTemplateFromItem(ctx context.Context, checklistId uint, name string, description *string, checklistItemId uint) (domain.Template, domain.Error) {
	// Guard rail: verify user has access to checklist
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return domain.Template{}, coreError.NewChecklistNotFoundError(checklistId)
	}

	// Fetch the specific checklist item
	item, findErr := service.checklistItemService.FindChecklistItemById(ctx, checklistId, checklistItemId)
	if findErr != nil {
		return domain.Template{}, findErr
	}
	if item == nil {
		return domain.Template{}, domain.NewError("Checklist item not found", 404)
	}

	// Build template rows from the item's rows
	rows := make([]domain.TemplateRow, len(item.Rows))
	for i, row := range item.Rows {
		rows[i] = domain.TemplateRow{
			Name:     row.Name,
			Position: float64(i+1) * domain.DefaultGapSize,
		}
	}

	template := domain.Template{
		Name:        name,
		Description: description,
		Rows:        rows,
	}

	return service.SaveTemplate(ctx, template)
}

func (service *templateService) ApplyTemplateToChecklist(ctx context.Context, checklistId uint, templateId uint) (domain.ChecklistItem, domain.Error) {
	// Guard rail: verify user has access to checklist
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return domain.ChecklistItem{}, coreError.NewChecklistNotFoundError(checklistId)
	}

	// Get template
	template, err := service.templateRepository.FindTemplateById(ctx, templateId)
	if err != nil {
		return domain.ChecklistItem{}, err
	}

	if template == nil {
		return domain.ChecklistItem{}, coreError.NewTemplateNotFoundError(templateId)
	}

	// Create one checklist item from the template
	newItem := domain.ChecklistItem{
		Name:      template.Name,
		Completed: false,
	}

	savedItem, domainErr := service.checklistItemService.SaveChecklistItem(ctx, checklistId, newItem)
	if domainErr != nil {
		return domain.ChecklistItem{}, domainErr
	}

	// Create rows for the item from template rows
	for _, templateRow := range template.Rows {
		row := domain.ChecklistItemRow{
			Name:      templateRow.Name,
			Completed: false,
		}
		savedRow, rowErr := service.checklistItemService.SaveChecklistItemRow(ctx, checklistId, savedItem.Id, row)
		if rowErr != nil {
			return domain.ChecklistItem{}, rowErr
		}
		savedItem.Rows = append(savedItem.Rows, savedRow)
	}

	return savedItem, nil
}

func (service *templateService) LeaveSharedTemplate(ctx context.Context, templateId uint) domain.Error {
	// Check user has access (not ownership — owner cannot leave their own template)
	if err := service.templateOwnershipChecker.HasAccessToTemplate(ctx, templateId); err != nil {
		return coreError.NewTemplateNotFoundError(templateId)
	}

	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return err
	}

	// Check if user is the owner — owners cannot leave
	isOwner, ownerErr := service.templateRepository.CheckUserIsTemplateOwner(ctx, templateId, userId)
	if ownerErr != nil {
		return ownerErr
	}
	if isOwner {
		return domain.NewError("Template owners cannot leave their own template", 400)
	}

	return service.templateRepository.DeleteTemplateShare(ctx, templateId, userId)
}

func CreateTemplateService(
	templateRepository repository.ITemplateRepository,
	templateOwnershipChecker guardrail.ITemplateOwnershipChecker,
	checklistItemService IChecklistItemsService,
	checklistOwnershipChecker guardrail.IChecklistOwnershipChecker,
) ITemplateService {
	return &templateService{
		templateRepository:        templateRepository,
		templateOwnershipChecker:  templateOwnershipChecker,
		checklistItemService:      checklistItemService,
		checklistOwnershipChecker: checklistOwnershipChecker,
	}
}
