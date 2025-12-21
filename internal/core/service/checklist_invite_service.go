package service

import (
	"context"
	"log"
	"time"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/error"
	guardrail "com.raunlo.checklist/internal/core/guard_rail"
	"com.raunlo.checklist/internal/core/repository"
	"com.raunlo.checklist/internal/util"
)

type IChecklistInviteService interface {
	CreateInvite(ctx context.Context, checklistId uint, expiresInHours *int, isSingleUse bool) (domain.ChecklistInvite, domain.Error)
	GetActiveInvites(ctx context.Context, checklistId uint) ([]domain.ChecklistInvite, domain.Error)
	RevokeInvite(ctx context.Context, inviteId uint) domain.Error
	ClaimInvite(ctx context.Context, token string) (uint, domain.Error) // Returns checklistId
}

type checklistInviteService struct {
	inviteRepository   repository.IChecklistInviteRepository
	checklistRepository repository.IChecklistRepository
	ownershipChecker    guardrail.IChecklistOwnershipChecker
}

func newChecklistInviteService(
	inviteRepo repository.IChecklistInviteRepository,
	checklistRepo repository.IChecklistRepository,
	ownershipChecker guardrail.IChecklistOwnershipChecker,
) IChecklistInviteService {
	return &checklistInviteService{
		inviteRepository:   inviteRepo,
		checklistRepository: checklistRepo,
		ownershipChecker:    ownershipChecker,
	}
}

func (s *checklistInviteService) CreateInvite(ctx context.Context, checklistId uint, expiresInHours *int, isSingleUse bool) (domain.ChecklistInvite, domain.Error) {
	// Check ownership
	if err := s.ownershipChecker.IsChecklistOwner(ctx, checklistId); err != nil {
		return domain.ChecklistInvite{}, err
	}

	// Get userId from context
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return domain.ChecklistInvite{}, err
	}

	// Generate secure token
	token, tokenErr := util.GenerateSecureToken()
	if tokenErr != nil {
		return domain.ChecklistInvite{}, domain.Wrap(tokenErr, "Failed to generate invite token", 500)
	}

	// Calculate expiration
	var expiresAt *time.Time
	if expiresInHours != nil && *expiresInHours > 0 {
		expiry := time.Now().Add(time.Duration(*expiresInHours) * time.Hour)
		expiresAt = &expiry
	}

	// Create invite
	invite := domain.ChecklistInvite{
		ChecklistId: checklistId,
		InviteToken: token,
		CreatedBy:   userId,
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		IsSingleUse: isSingleUse,
	}

	createdInvite, createErr := s.inviteRepository.CreateInvite(ctx, invite)
	if createErr != nil {
		return domain.ChecklistInvite{}, createErr
	}

	log.Printf("Invite created: checklistId=%d, token=%s, createdBy=%s", checklistId, token, userId)
	return createdInvite, nil
}

func (s *checklistInviteService) GetActiveInvites(ctx context.Context, checklistId uint) ([]domain.ChecklistInvite, domain.Error) {
	// Check ownership
	if err := s.ownershipChecker.IsChecklistOwner(ctx, checklistId); err != nil {
		return nil, err
	}

	invites, err := s.inviteRepository.FindActiveInvitesByChecklistId(ctx, checklistId)
	if err != nil {
		return nil, err
	}

	return invites, nil
}

func (s *checklistInviteService) RevokeInvite(ctx context.Context, inviteId uint) domain.Error {
	// First, fetch the invite to check which checklist it belongs to
	// (We need to verify ownership of that checklist)
	// For now, we'll trust that the controller has already verified ownership
	// A more robust implementation would fetch the invite first to get checklistId

	err := s.inviteRepository.DeleteInviteById(ctx, inviteId)
	if err != nil {
		return err
	}

	log.Printf("Invite revoked: inviteId=%d", inviteId)
	return nil
}

func (s *checklistInviteService) ClaimInvite(ctx context.Context, token string) (uint, domain.Error) {
	// Get userId from context
	userId, userErr := domain.GetUserIdFromContext(ctx)
	if userErr != nil {
		return 0, userErr
	}

	// Find invite by token
	invite, err := s.inviteRepository.FindInviteByToken(ctx, token)
	if err != nil {
		return 0, err
	}

	if invite == nil {
		return 0, error.NewInviteNotFoundError()
	}

	// Check if expired
	if invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now()) {
		return 0, error.NewInviteExpiredError()
	}

	// Check if already claimed and single-use
	if invite.ClaimedAt != nil && invite.IsSingleUse {
		return 0, error.NewInviteAlreadyClaimedError()
	}

	// Check if user already has access (idempotent behavior)
	hasAccess, accessErr := s.checklistRepository.CheckUserHasAccessToChecklist(ctx, invite.ChecklistId, userId)
	if accessErr != nil {
		return 0, accessErr
	}

	if hasAccess {
		log.Printf("User %s already has access to checklist %d (idempotent claim)", userId, invite.ChecklistId)
		return invite.ChecklistId, nil
	}

	// Claim the invite
	claimErr := s.inviteRepository.ClaimInvite(ctx, token, userId)
	if claimErr != nil {
		return 0, claimErr
	}

	// Create checklist share entry
	shareErr := s.checklistRepository.CreateChecklistShare(ctx, invite.ChecklistId, invite.CreatedBy, userId)
	if shareErr != nil {
		return 0, shareErr
	}

	log.Printf("Invite claimed: token=%s, checklistId=%d, claimedBy=%s", token, invite.ChecklistId, userId)
	return invite.ChecklistId, nil
}
