package guardrail

import (
	"context"
	"log"

	"com.raunlo.checklist/internal/core/domain"
	domainErr "com.raunlo.checklist/internal/core/error"
	"com.raunlo.checklist/internal/core/repository"
)

type IWorkspaceOwnershipChecker interface {
	IsWorkspaceOwner(ctx context.Context, workspaceId uint) domain.Error
	IsMember(ctx context.Context, workspaceId uint) domain.Error
}

type workspaceOwnershipCheckerService struct {
	repository repository.IWorkspaceRepository
}

func (service *workspaceOwnershipCheckerService) IsWorkspaceOwner(ctx context.Context, workspaceId uint) domain.Error {
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return err
	}

	isOwner, err := service.repository.CheckUserIsWorkspaceOwner(ctx, workspaceId, userId)
	log.Printf("GuardRail: User(id=%s) owner check for workspace %d: %v", domain.GetHashedUserIdFromContext(ctx), workspaceId, isOwner)
	if err != nil {
		return domain.Wrap(err, "Failed to check workspace ownership", 500)
	}
	if !isOwner {
		return domainErr.NewWorkspaceNotFoundError(workspaceId)
	}

	return nil
}

func (service *workspaceOwnershipCheckerService) IsMember(ctx context.Context, workspaceId uint) domain.Error {
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return err
	}

	isMember, err := service.repository.CheckUserIsMember(ctx, workspaceId, userId)
	log.Printf("GuardRail: User(id=%s) member check for workspace %d: %v", domain.GetHashedUserIdFromContext(ctx), workspaceId, isMember)
	if err != nil {
		return domain.Wrap(err, "Failed to check workspace membership", 500)
	}
	if !isMember {
		return domainErr.NewWorkspaceNotFoundError(workspaceId)
	}

	return nil
}

func NewWorkspaceOwnershipCheckerService(repository repository.IWorkspaceRepository) IWorkspaceOwnershipChecker {
	return &workspaceOwnershipCheckerService{
		repository: repository,
	}
}
