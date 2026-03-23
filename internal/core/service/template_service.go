package service

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/error"
	guardrail "com.raunlo.checklist/internal/core/guard_rail"
	"com.raunlo.checklist/internal/core/notification"
	"com.raunlo.checklist/internal/core/repository"
)

type ITemplateService interface {
	SaveTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error)
	FindTemplateById(ctx context.Context, id uint) (*domain.Template, domain.Error)
	FindAllTemplates(ctx context.Context) ([]domain.Template, domain.Error)
	UpdateTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error)
	DeleteTemplate(ctx context.Context, id uint) domain.Error
	CreateTemplateFromItems(ctx context.Context, checklistId uint, name string, description *string, checklistItemIds []uint) (domain.Template, domain.Error)
	ApplyTemplateToChecklist(ctx context.Context, checklistId uint, templateId uint, selectedItemIds []uint) ([]domain.ChecklistItem, domain.Error)
	CreateChecklistFromTemplate(ctx context.Context, templateId uint, checklistName string) (*domain.Checklist, domain.Error)
	GetTemplatePreview(ctx context.Context, checklistId uint, templateId uint) ([]domain.TemplateItem, []domain.TemplateItem, domain.Error)
}

type templateService struct {
	templateRepository        repository.ITemplateRepository
	templateOwnershipChecker  guardrail.ITemplateOwnershipChecker
	checklistRepository       repository.IChecklistRepository
	checklistItemService      IChecklistItemsService
	checklistOwnershipChecker guardrail.IChecklistOwnershipChecker
	notificationService       notification.INotificationService
}

func (service *templateService) SaveTemplate(ctx context.Context, template domain.Template) (domain.Template, domain.Error) {
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return domain.Template{}, err
	}

	template.UserID = userId
	return service.templateRepository.SaveTemplate(ctx, template)
}

func (service *templateService) FindTemplateById(ctx context.Context, id uint) (*domain.Template, domain.Error) {
	if err := service.templateOwnershipChecker.IsTemplateOwner(ctx, id); err != nil {
		return nil, error.NewTemplateNotFoundError(id)
	}
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
	if err := service.templateOwnershipChecker.IsTemplateOwner(ctx, template.ID); err != nil {
		return domain.Template{}, error.NewTemplateNotFoundError(template.ID)
	}
	return service.templateRepository.UpdateTemplate(ctx, template)
}

func (service *templateService) DeleteTemplate(ctx context.Context, id uint) domain.Error {
	if err := service.templateOwnershipChecker.IsTemplateOwner(ctx, id); err != nil {
		return error.NewTemplateNotFoundError(id)
	}
	return service.templateRepository.DeleteTemplate(ctx, id)
}

func (service *templateService) CreateTemplateFromItems(ctx context.Context, checklistId uint, name string, description *string, checklistItemIds []uint) (domain.Template, domain.Error) {
	// Guard rail: verify user owns the checklist
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return domain.Template{}, error.NewChecklistNotFoundError(checklistId)
	}

	// Fetch all items in the checklist
	checklistItems, err := service.checklistItemService.FindAllChecklistItems(ctx, checklistId, nil, domain.AscSort)
	if err != nil {
		return domain.Template{}, err
	}

	// Build a map of requested IDs for quick lookup
	requestedIds := make(map[uint]bool)
	for _, id := range checklistItemIds {
		requestedIds[id] = true
	}

	// Filter checklist items to only those requested
	var selectedItems []domain.ChecklistItem
	for _, item := range checklistItems {
		if requestedIds[item.Id] {
			selectedItems = append(selectedItems, item)
		}
	}

	// Validate we found at least one matching item
	if len(selectedItems) == 0 {
		return domain.Template{}, domain.NewError("No valid checklist items found for given IDs", 400)
	}

	// Convert checklist items to template items
	var templateItems []domain.TemplateItem
	for i, item := range selectedItems {
		templateItems = append(templateItems, domain.TemplateItem{
			Name:     item.Name,
			Position: float64(i) * domain.DefaultGapSize,
		})
	}

	// Build the template
	template := domain.Template{
		Name:        name,
		Description: description,
		Items:       templateItems,
	}

	// Save the template (SaveTemplate handles setting UserID from context)
	return service.SaveTemplate(ctx, template)
}

func (service *templateService) ApplyTemplateToChecklist(ctx context.Context, checklistId uint, templateId uint, selectedItemIds []uint) ([]domain.ChecklistItem, domain.Error) {
	// Guard rails
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return nil, error.NewChecklistNotFoundError(checklistId)
	}

	if err := service.templateOwnershipChecker.IsTemplateOwner(ctx, templateId); err != nil {
		return nil, error.NewTemplateNotFoundError(templateId)
	}

	// Get template
	template, err := service.templateRepository.FindTemplateById(ctx, templateId)
	if err != nil {
		return nil, err
	}

	if template == nil {
		return nil, error.NewTemplateNotFoundError(templateId)
	}

	// Add selected items to checklist
	createdItems := make([]domain.ChecklistItem, 0)
	for _, itemId := range selectedItemIds {
		// Find template item by ID
		var selectedItem *domain.TemplateItem
		for _, ti := range template.Items {
			if ti.ID == itemId {
				selectedItem = &ti
				break
			}
		}

		if selectedItem == nil {
			continue
		}

		// Create checklist item
		checklistItem := domain.ChecklistItem{
			Name:      selectedItem.Name,
			Completed: false,
		}

		created, domainErr := service.checklistItemService.SaveChecklistItem(ctx, checklistId, checklistItem)
		if domainErr != nil {
			return nil, domainErr
		}

		createdItems = append(createdItems, created)
	}

	return createdItems, nil
}

func (service *templateService) CreateChecklistFromTemplate(ctx context.Context, templateId uint, checklistName string) (*domain.Checklist, domain.Error) {
	// Guard rail
	if err := service.templateOwnershipChecker.IsTemplateOwner(ctx, templateId); err != nil {
		return nil, error.NewTemplateNotFoundError(templateId)
	}

	// Get template
	template, err := service.templateRepository.FindTemplateById(ctx, templateId)
	if err != nil {
		return nil, err
	}

	if template == nil {
		return nil, error.NewTemplateNotFoundError(templateId)
	}

	// Create checklist
	checklist := domain.Checklist{
		Name: checklistName,
	}

	created, domainErr := service.checklistRepository.SaveChecklist(ctx, checklist)
	if domainErr != nil {
		return nil, domainErr
	}

	// Add template items to checklist
	for i, item := range template.Items {
		checklistItem := domain.ChecklistItem{
			Name:      item.Name,
			Completed: false,
			Position:  float64(i) * domain.DefaultGapSize,
		}

		_, domainErr := service.checklistItemService.SaveChecklistItem(ctx, created.Id, checklistItem)
		if domainErr != nil {
			return nil, domainErr
		}
	}

	return &created, nil
}

func (service *templateService) GetTemplatePreview(ctx context.Context, checklistId uint, templateId uint) ([]domain.TemplateItem, []domain.TemplateItem, domain.Error) {
	// Guard rails
	if err := service.checklistOwnershipChecker.HasAccessToChecklist(ctx, checklistId); err != nil {
		return nil, nil, error.NewChecklistNotFoundError(checklistId)
	}

	if err := service.templateOwnershipChecker.IsTemplateOwner(ctx, templateId); err != nil {
		return nil, nil, error.NewTemplateNotFoundError(templateId)
	}

	// Get template
	template, err := service.templateRepository.FindTemplateById(ctx, templateId)
	if err != nil {
		return nil, nil, err
	}

	if template == nil {
		return nil, nil, error.NewTemplateNotFoundError(templateId)
	}

	// Get checklist items
	checklistItems, err := service.checklistItemService.FindAllChecklistItems(ctx, checklistId, nil, domain.AscSort)
	if err != nil {
		return nil, nil, err
	}

	// Build set of existing item names (lowercase)
	existingNames := make(map[string]bool)
	for _, item := range checklistItems {
		existingNames[toLowercase(item.Name)] = true
	}

	// Identify duplicates and new items
	newItems := make([]domain.TemplateItem, 0)
	existingItems := make([]domain.TemplateItem, 0)

	for _, item := range template.Items {
		if existingNames[toLowercase(item.Name)] {
			existingItems = append(existingItems, item)
		} else {
			newItems = append(newItems, item)
		}
	}

	return existingItems, newItems, nil
}

func toLowercase(s string) string {
	// Simple lowercase conversion for duplicate detection
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

func CreateTemplateService(
	templateRepository repository.ITemplateRepository,
	templateOwnershipChecker guardrail.ITemplateOwnershipChecker,
	checklistRepository repository.IChecklistRepository,
	checklistItemService IChecklistItemsService,
	checklistOwnershipChecker guardrail.IChecklistOwnershipChecker,
	notificationService notification.INotificationService,
) ITemplateService {
	return &templateService{
		templateRepository:       templateRepository,
		templateOwnershipChecker: templateOwnershipChecker,
		checklistRepository:      checklistRepository,
		checklistItemService:     checklistItemService,
		checklistOwnershipChecker: checklistOwnershipChecker,
		notificationService:      notificationService,
	}
}
