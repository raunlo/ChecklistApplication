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

type ITemplateInviteService interface {
	CreateInvite(ctx context.Context, templateId uint, name *string, expiresInHours *int, isSingleUse bool) (domain.TemplateInvite, domain.Error)
	GetActiveInvites(ctx context.Context, templateId uint) ([]domain.TemplateInvite, domain.Error)
	RevokeInvite(ctx context.Context, inviteId uint) domain.Error
	ClaimInvite(ctx context.Context, token string) (uint, domain.Error) // Returns templateId
}

type templateInviteService struct {
	inviteRepository   repository.ITemplateInviteRepository
	templateRepository repository.ITemplateRepository
	ownershipChecker   guardrail.ITemplateOwnershipChecker
}

func NewTemplateInviteService(
	inviteRepo repository.ITemplateInviteRepository,
	templateRepo repository.ITemplateRepository,
	ownershipChecker guardrail.ITemplateOwnershipChecker,
) ITemplateInviteService {
	return &templateInviteService{
		inviteRepository:   inviteRepo,
		templateRepository: templateRepo,
		ownershipChecker:   ownershipChecker,
	}
}

func (s *templateInviteService) CreateInvite(ctx context.Context, templateId uint, name *string, expiresInHours *int, isSingleUse bool) (domain.TemplateInvite, domain.Error) {
	// Check ownership
	if err := s.ownershipChecker.IsTemplateOwner(ctx, templateId); err != nil {
		return domain.TemplateInvite{}, err
	}

	// Get userId from context
	userId, err := domain.GetUserIdFromContext(ctx)
	if err != nil {
		return domain.TemplateInvite{}, err
	}

	// Check for existing active invites to prevent too many
	existingInvites, listErr := s.inviteRepository.FindActiveInvitesByTemplateId(ctx, templateId)
	if listErr != nil {
		return domain.TemplateInvite{}, listErr
	}

	// Limit to max 10 active invites per template to prevent abuse
	if len(existingInvites) >= 10 {
		return domain.TemplateInvite{}, domain.NewError("Maximum number of active invites (10) reached for this template. Please revoke old invites first.", 400)
	}

	// Generate secure token
	token, tokenErr := util.GenerateSecureToken()
	if tokenErr != nil {
		return domain.TemplateInvite{}, domain.Wrap(tokenErr, "Failed to generate invite token", 500)
	}

	// Calculate expiration with timezone awareness
	var expiresAt *time.Time
	if expiresInHours != nil && *expiresInHours > 0 {
		expiry := time.Now().UTC().Add(time.Duration(*expiresInHours) * time.Hour)
		expiresAt = &expiry
	}

	invite := domain.TemplateInvite{
		TemplateId:  templateId,
		Name:        name,
		InviteToken: token,
		CreatedBy:   userId,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   expiresAt,
		IsSingleUse: isSingleUse,
	}

	createdInvite, createErr := s.inviteRepository.CreateInvite(ctx, invite)
	if createErr != nil {
		return domain.TemplateInvite{}, createErr
	}

	log.Printf("Template invite created: templateId=%d, name=%v, token=%s..., createdBy=%s", templateId, name, token[:8], domain.GetHashedUserIdFromContext(ctx))
	return createdInvite, nil
}

func (s *templateInviteService) GetActiveInvites(ctx context.Context, templateId uint) ([]domain.TemplateInvite, domain.Error) {
	// Check ownership
	if err := s.ownershipChecker.IsTemplateOwner(ctx, templateId); err != nil {
		return nil, err
	}

	invites, err := s.inviteRepository.FindActiveInvitesByTemplateId(ctx, templateId)
	if err != nil {
		return nil, err
	}

	return invites, nil
}

func (s *templateInviteService) RevokeInvite(ctx context.Context, inviteId uint) domain.Error {
	err := s.inviteRepository.DeleteInviteById(ctx, inviteId)
	if err != nil {
		return err
	}

	log.Printf("Template invite revoked: inviteId=%d", inviteId)
	return nil
}

func (s *templateInviteService) ClaimInvite(ctx context.Context, token string) (uint, domain.Error) {
	// Input validation
	if token == "" {
		return 0, domain.NewError("Invite token is required", 400)
	}

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

	// Check if expired (timezone-aware comparison)
	if invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now().UTC()) {
		return 0, error.NewInviteExpiredError()
	}

	// Check if already claimed and single-use
	if invite.ClaimedAt != nil && invite.IsSingleUse {
		return 0, error.NewInviteAlreadyClaimedError()
	}

	// Check if user already has access (idempotent behavior)
	hasAccess, accessErr := s.templateRepository.CheckUserHasAccessToTemplate(ctx, invite.TemplateId, userId)
	if accessErr != nil {
		return 0, accessErr
	}

	if hasAccess {
		log.Printf("User %s already has access to template %d (idempotent claim)", domain.GetHashedUserIdFromContext(ctx), invite.TemplateId)
		return invite.TemplateId, nil
	}

	// Claim the invite and create share in a single transaction
	claimAndShareErr := s.inviteRepository.ClaimInviteAndCreateShare(ctx, token, userId, invite.TemplateId, invite.CreatedBy)
	if claimAndShareErr != nil {
		return 0, claimAndShareErr
	}

	log.Printf("Template invite claimed: token=%s..., templateId=%d, claimedBy=%s", token[:8], invite.TemplateId, domain.GetHashedUserIdFromContext(ctx))
	return invite.TemplateId, nil
}
