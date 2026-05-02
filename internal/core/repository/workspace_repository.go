package repository

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
)

type IWorkspaceRepository interface {
	SaveWorkspace(ctx context.Context, workspace domain.Workspace) (domain.Workspace, domain.Error)
	FindWorkspaceById(ctx context.Context, id uint) (*domain.Workspace, domain.Error)
	FindWorkspacesByUserId(ctx context.Context, userId string) ([]domain.Workspace, domain.Error)
	UpdateWorkspace(ctx context.Context, workspace domain.Workspace) (domain.Workspace, domain.Error)
	DeleteWorkspace(ctx context.Context, id uint) domain.Error
	CheckUserIsWorkspaceOwner(ctx context.Context, workspaceId uint, userId string) (bool, domain.Error)
	CheckUserIsMember(ctx context.Context, workspaceId uint, userId string) (bool, domain.Error)
	GetWorkspaceMembers(ctx context.Context, workspaceId uint) ([]domain.WorkspaceMember, domain.Error)
	RemoveMember(ctx context.Context, workspaceId uint, userId string) domain.Error
	AddMember(ctx context.Context, workspaceId uint, userId string) domain.Error
	FindDefaultWorkspace(ctx context.Context, userId string) (*domain.Workspace, domain.Error)
}

type IWorkspaceInviteRepository interface {
	CreateInvite(ctx context.Context, invite domain.WorkspaceInvite) (domain.WorkspaceInvite, domain.Error)
	FindInviteByToken(ctx context.Context, token string) (*domain.WorkspaceInvite, domain.Error)
	FindActiveInvitesByWorkspaceId(ctx context.Context, workspaceId uint) ([]domain.WorkspaceInvite, domain.Error)
	DeleteInviteById(ctx context.Context, inviteId uint) domain.Error
	ClaimInviteAndAddMember(ctx context.Context, token string, userId string, workspaceId uint, isSingleUse bool) domain.Error
}
