package service

import (
	"com.raunlo.checklist/internal/core/notification"
	"com.raunlo.checklist/internal/core/repository"
)

// ServiceFactory holds all services
type ServiceFactory struct {
}

// NewServiceFactory creates a new service factory
func NewServiceFactory() *ServiceFactory {
	// FCM removed from service factory
	return &ServiceFactory{}
}

// GetFCMService returns the FCM service
// FCM removed

func CreateChecklistService(checklistRepository repository.IChecklistRepository) IChecklistService {
	return &checklistService{
		repository: checklistRepository,
	}
}

func CreateChecklistItemService(repository repository.IChecklistItemsRepository,
	notificationService notification.INotificationService) IChecklistItemsService {
	return &checklistItemsService{
		repository: repository,
		notifier:   notificationService,
	}
}

func CreateChecklistItemTemplateService(repository repository.IChecklistItemTemplateRepository) IChecklistItemTemplateService {
	return &checklistItemTemplateService{
		repository: repository,
	}
}
