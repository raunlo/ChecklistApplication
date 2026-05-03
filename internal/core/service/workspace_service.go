package service

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	domainErr "com.raunlo.checklist/internal/core/error"
	guardrail "com.raunlo.checklist/internal/core/guard_rail"
	"com.raunlo.checklist/internal/core/repository"
)

type IWorkspaceService interface {
	CreateWorkspace(ctx context.Context, workspace domain.Workspace) (domain.Workspace, domain.Error)
	FindWorkspaceById(ctx context.Context, id uint) (*domain.Workspace, domain.Error)
	FindAllWorkspaces(ctx context.Context) ([]domain.Workspace, domain.Error)
	UpdateWorkspace(ctx context.Context, workspace domain.Workspace) (domain.Workspace, domain.Error)
	DeleteWorkspace(ctx context.Context, id uint) domain.Error
	GetMembers(ctx context.Context, workspaceId uint) ([]domain.WorkspaceMember, domain.Error)
	RemoveMember(ctx context.Context, workspaceId uint, memberId uint) domain.Error
	LeaveWorkspace(ctx context.Context, workspaceId uint) domain.Error
	GetWorkspaceTemplates(ctx context.Context, workspaceId uint) ([]domain.Template, domain.Error)
	GetWorkspaceChecklists(ctx context.Context, workspaceId uint) ([]domain.Checklist, domain.Error)
	HasDefaultWorkspace(ctx context.Context, userId string) (bool, domain.Error)
	FindDefaultWorkspace(ctx context.Context, userId string) (*domain.Workspace, domain.Error)
}

type workspaceService struct {
	workspaceRepository repository.IWorkspaceRepository
	checklistRepository repository.IChecklistRepository
	templateRepository  repository.ITemplateRepository
	ownershipChecker    guardrail.IWorkspaceOwnershipChecker
}

func (s *workspaceService) CreateWorkspace(ctx context.Context, workspace domain.Workspace) (domain.Workspace, domain.Error) {
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return domain.Workspace{}, err
	}

	workspace.OwnerUserId = userId
	created, err := s.workspaceRepository.SaveWorkspace(ctx, workspace)
	if err != nil {
		return domain.Workspace{}, err
	}

	// Add owner as a member
	if addErr := s.workspaceRepository.AddMember(ctx, created.Id, userId); addErr != nil {
		return domain.Workspace{}, addErr
	}

	created.IsOwner = true
	created.MemberCount = 1
	return created, nil
}

func (s *workspaceService) FindWorkspaceById(ctx context.Context, id uint) (*domain.Workspace, domain.Error) {
	if err := s.ownershipChecker.IsMember(ctx, id); err != nil {
		return nil, err
	}
	return s.workspaceRepository.FindWorkspaceById(ctx, id)
}

func (s *workspaceService) FindAllWorkspaces(ctx context.Context) ([]domain.Workspace, domain.Error) {
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.workspaceRepository.FindWorkspacesByUserId(ctx, userId)
}

func (s *workspaceService) UpdateWorkspace(ctx context.Context, workspace domain.Workspace) (domain.Workspace, domain.Error) {
	if err := s.ownershipChecker.IsWorkspaceOwner(ctx, workspace.Id); err != nil {
		return domain.Workspace{}, err
	}
	return s.workspaceRepository.UpdateWorkspace(ctx, workspace)
}

func (s *workspaceService) DeleteWorkspace(ctx context.Context, id uint) domain.Error {
	if err := s.ownershipChecker.IsWorkspaceOwner(ctx, id); err != nil {
		return err
	}
	ws, err := s.workspaceRepository.FindWorkspaceById(ctx, id)
	if err != nil {
		return err
	}
	if ws != nil && ws.IsDefault {
		return domain.NewError("Cannot delete the default personal workspace", 400)
	}
	return s.workspaceRepository.DeleteWorkspace(ctx, id)
}

func (s *workspaceService) GetMembers(ctx context.Context, workspaceId uint) ([]domain.WorkspaceMember, domain.Error) {
	if err := s.ownershipChecker.IsMember(ctx, workspaceId); err != nil {
		return nil, err
	}
	return s.workspaceRepository.GetWorkspaceMembers(ctx, workspaceId)
}

func (s *workspaceService) RemoveMember(ctx context.Context, workspaceId uint, memberId uint) domain.Error {
	if err := s.ownershipChecker.IsWorkspaceOwner(ctx, workspaceId); err != nil {
		return err
	}
	return s.workspaceRepository.RemoveMember(ctx, workspaceId, memberId)
}

func (s *workspaceService) LeaveWorkspace(ctx context.Context, workspaceId uint) domain.Error {
	if err := s.ownershipChecker.IsMember(ctx, workspaceId); err != nil {
		return domainErr.NewWorkspaceNotFoundError(workspaceId)
	}

	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return err
	}

	isOwner, err := s.workspaceRepository.CheckUserIsWorkspaceOwner(ctx, workspaceId, userId)
	if err != nil {
		return err
	}
	if isOwner {
		return domain.NewError("Workspace owners cannot leave. Delete the workspace instead.", 400)
	}

	return s.workspaceRepository.RemoveSelf(ctx, workspaceId, userId)
}

func (s *workspaceService) GetWorkspaceTemplates(ctx context.Context, workspaceId uint) ([]domain.Template, domain.Error) {
	if err := s.ownershipChecker.IsMember(ctx, workspaceId); err != nil {
		return nil, err
	}
	return s.templateRepository.FindTemplatesByWorkspaceId(ctx, workspaceId)
}

func (s *workspaceService) GetWorkspaceChecklists(ctx context.Context, workspaceId uint) ([]domain.Checklist, domain.Error) {
	if err := s.ownershipChecker.IsMember(ctx, workspaceId); err != nil {
		return nil, err
	}
	return s.checklistRepository.FindChecklistsByWorkspaceId(ctx, workspaceId)
}

func (s *workspaceService) HasDefaultWorkspace(ctx context.Context, userId string) (bool, domain.Error) {
	ws, err := s.workspaceRepository.FindDefaultWorkspace(ctx, userId)
	if err != nil {
		return false, err
	}
	return ws != nil, nil
}

func (s *workspaceService) FindDefaultWorkspace(ctx context.Context, userId string) (*domain.Workspace, domain.Error) {
	return s.workspaceRepository.FindDefaultWorkspace(ctx, userId)
}

func CreateWorkspaceService(
	workspaceRepository repository.IWorkspaceRepository,
	checklistRepository repository.IChecklistRepository,
	templateRepository repository.ITemplateRepository,
	ownershipChecker guardrail.IWorkspaceOwnershipChecker,
) IWorkspaceService {
	return &workspaceService{
		workspaceRepository: workspaceRepository,
		checklistRepository: checklistRepository,
		templateRepository:  templateRepository,
		ownershipChecker:    ownershipChecker,
	}
}
