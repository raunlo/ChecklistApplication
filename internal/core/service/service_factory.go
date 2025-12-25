package service

import (
	guardrail "com.raunlo.checklist/internal/core/guard_rail"
	"com.raunlo.checklist/internal/core/notification"
	"com.raunlo.checklist/internal/core/repository"
)

// ServiceFactory holds all services
type ServiceFactory struct{}

// NewServiceFactory creates a new service factory
func NewServiceFactory() *ServiceFactory {
	// FCM removed from service factory
	return &ServiceFactory{}
}

// GetFCMService returns the FCM service
// FCM removed

func CreateChecklistService(checklistRepository repository.IChecklistRepository,
	checklistOwnershipChecker guardrail.IChecklistOwnershipChecker,
	checklistItemService IChecklistItemsService) IChecklistService {
	return &checklistService{
		repository:                checklistRepository,
		checklistOwnershipChecker: checklistOwnershipChecker,
		checklistItemService:      checklistItemService,
	}
}

func CreateChecklistItemService(repository repository.IChecklistItemsRepository,
	notificationService notification.INotificationService,
	checklistOwnershipChecker guardrail.IChecklistOwnershipChecker,
) IChecklistItemsService {
	return &checklistItemsService{
		repository:                repository,
		notifier:                  notificationService,
		checklistOwnershipChecker: checklistOwnershipChecker,
	}
}

func CreateChecklistItemTemplateService(repository repository.IChecklistItemTemplateRepository) IChecklistItemTemplateService {
	return &checklistItemTemplateService{
		repository: repository,
	}
}

func CreateChecklistInviteService(
	inviteRepo repository.IChecklistInviteRepository,
	checklistRepo repository.IChecklistRepository,
	ownershipChecker guardrail.IChecklistOwnershipChecker,
) IChecklistInviteService {
	return newChecklistInviteService(inviteRepo, checklistRepo, ownershipChecker)
}
