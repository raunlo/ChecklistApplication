package service

import (
	"context"
	"log"
	"time"

	"com.raunlo.checklist/internal/core/domain"
	domainError "com.raunlo.checklist/internal/core/error"
	guardrail "com.raunlo.checklist/internal/core/guard_rail"
	"com.raunlo.checklist/internal/core/repository"
	"com.raunlo.checklist/internal/util"
)

type IWorkspaceInviteService interface {
	CreateInvite(ctx context.Context, workspaceId uint, name *string, expiresInHours *int, isSingleUse bool) (domain.WorkspaceInvite, domain.Error)
	GetActiveInvites(ctx context.Context, workspaceId uint) ([]domain.WorkspaceInvite, domain.Error)
	RevokeInvite(ctx context.Context, workspaceId uint, inviteId uint) domain.Error
	ClaimInvite(ctx context.Context, token string) (uint, domain.Error) // returns workspaceId
}

type workspaceInviteService struct {
	inviteRepository    repository.IWorkspaceInviteRepository
	workspaceRepository repository.IWorkspaceRepository
	ownershipChecker    guardrail.IWorkspaceOwnershipChecker
}

func (s *workspaceInviteService) CreateInvite(ctx context.Context, workspaceId uint, name *string, expiresInHours *int, isSingleUse bool) (domain.WorkspaceInvite, domain.Error) {
	if err := s.ownershipChecker.IsWorkspaceOwner(ctx, workspaceId); err != nil {
		return domain.WorkspaceInvite{}, err
	}

	ws, wsErr := s.workspaceRepository.FindWorkspaceById(ctx, workspaceId)
	if wsErr != nil {
		return domain.WorkspaceInvite{}, wsErr
	}
	if ws != nil && ws.IsDefault {
		return domain.WorkspaceInvite{}, domain.NewError("Cannot invite members to a personal workspace", 400)
	}
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return domain.WorkspaceInvite{}, err
	}

	existing, listErr := s.inviteRepository.FindActiveInvitesByWorkspaceId(ctx, workspaceId)
	if listErr != nil {
		return domain.WorkspaceInvite{}, listErr
	}
	if len(existing) >= 10 {
		return domain.WorkspaceInvite{}, domain.NewError("Maximum number of active invites (10) reached for this workspace.", 400)
	}

	token, tokenErr := util.GenerateSecureToken()
	if tokenErr != nil {
		return domain.WorkspaceInvite{}, domain.Wrap(tokenErr, "Failed to generate invite token", 500)
	}

	var expiresAt *time.Time
	if expiresInHours != nil && *expiresInHours > 0 {
		expiry := time.Now().UTC().Add(time.Duration(*expiresInHours) * time.Hour)
		expiresAt = &expiry
	}

	invite := domain.WorkspaceInvite{
		WorkspaceId: workspaceId,
		Name:        name,
		InviteToken: token,
		CreatedBy:   userId,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   expiresAt,
		IsSingleUse: isSingleUse,
	}

	created, createErr := s.inviteRepository.CreateInvite(ctx, invite)
	if createErr != nil {
		return domain.WorkspaceInvite{}, createErr
	}

	log.Printf("Workspace invite created: workspaceId=%d, token=%s..., createdBy=%s", workspaceId, token[:8], domain.GetHashedUserIdFromContext(ctx))
	return created, nil
}

func (s *workspaceInviteService) GetActiveInvites(ctx context.Context, workspaceId uint) ([]domain.WorkspaceInvite, domain.Error) {
	if err := s.ownershipChecker.IsWorkspaceOwner(ctx, workspaceId); err != nil {
		return nil, err
	}
	return s.inviteRepository.FindActiveInvitesByWorkspaceId(ctx, workspaceId)
}

func (s *workspaceInviteService) RevokeInvite(ctx context.Context, workspaceId uint, inviteId uint) domain.Error {
	if err := s.ownershipChecker.IsWorkspaceOwner(ctx, workspaceId); err != nil {
		return err
	}
	err := s.inviteRepository.DeleteInviteById(ctx, inviteId)
	if err != nil {
		return err
	}
	log.Printf("Workspace invite revoked: inviteId=%d", inviteId)
	return nil
}

func (s *workspaceInviteService) ClaimInvite(ctx context.Context, token string) (uint, domain.Error) {
	if token == "" {
		return 0, domain.NewError("Invite token is required", 400)
	}

	userId, userErr := domain.GetUserIdFromContext(ctx)
	if userErr != nil {
		return 0, userErr
	}

	invite, err := s.inviteRepository.FindInviteByToken(ctx, token)
	if err != nil {
		return 0, err
	}
	if invite == nil {
		return 0, domainError.NewInviteNotFoundError()
	}

	// Block claiming invites for default (personal) workspaces
	ws, wsErr := s.workspaceRepository.FindWorkspaceById(ctx, invite.WorkspaceId)
	if wsErr != nil {
		return 0, wsErr
	}
	if ws != nil && ws.IsDefault {
		return 0, domain.NewError("Cannot join a personal workspace", 400)
	}
	if invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now().UTC()) {
		return 0, domainError.NewInviteExpiredError()
	}

	// Idempotent: already a member — check FIRST before anything else
	isMember, memberErr := s.workspaceRepository.CheckUserIsMember(ctx, invite.WorkspaceId, userId)
	if memberErr != nil {
		return 0, memberErr
	}
	if isMember {
		log.Printf("User %s already member of workspace %d (idempotent claim)", domain.GetHashedUserIdFromContext(ctx), invite.WorkspaceId)
		return invite.WorkspaceId, nil
	}

	if invite.ClaimedAt != nil && invite.IsSingleUse {
		return 0, domainError.NewInviteAlreadyClaimedError()
	}

	if claimErr := s.inviteRepository.ClaimInviteAndAddMember(ctx, token, userId, invite.WorkspaceId, invite.IsSingleUse); claimErr != nil {
		return 0, claimErr
	}

	log.Printf("Workspace invite claimed: token=%s..., workspaceId=%d, claimedBy=%s", token[:8], invite.WorkspaceId, domain.GetHashedUserIdFromContext(ctx))
	return invite.WorkspaceId, nil
}

func CreateWorkspaceInviteService(
	inviteRepository repository.IWorkspaceInviteRepository,
	workspaceRepository repository.IWorkspaceRepository,
	ownershipChecker guardrail.IWorkspaceOwnershipChecker,
) IWorkspaceInviteService {
	return &workspaceInviteService{
		inviteRepository:    inviteRepository,
		workspaceRepository: workspaceRepository,
		ownershipChecker:    ownershipChecker,
	}
}
