package checklist

import (
	"time"

	"com.raunlo.checklist/internal/core/domain"
)

type IChecklistInviteDtoMapper interface {
	ToDTO(invite domain.ChecklistInvite, baseUrl string) InviteResponse
	ToDTOArray(invites []domain.ChecklistInvite, baseUrl string) []InviteResponse
}

type checklistInviteDtoMapper struct{}

func NewChecklistInviteDtoMapper() IChecklistInviteDtoMapper {
	return &checklistInviteDtoMapper{}
}

func (m *checklistInviteDtoMapper) ToDTO(invite domain.ChecklistInvite, baseUrl string) InviteResponse {
	// Calculate isExpired
	isExpired := false
	if invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now()) {
		isExpired = true
	}

	// Calculate isClaimed
	isClaimed := invite.ClaimedAt != nil

	// Build invite URL
	inviteUrl := baseUrl + "/invites/" + invite.InviteToken + "/claim"

	return InviteResponse{
		Id:          invite.Id,
		ChecklistId: invite.ChecklistId,
		Name:        invite.Name,
		InviteToken: invite.InviteToken,
		InviteUrl:   inviteUrl,
		CreatedAt:   invite.CreatedAt,
		ExpiresAt:   invite.ExpiresAt,
		ClaimedBy:   invite.ClaimedBy,
		ClaimedAt:   invite.ClaimedAt,
		IsSingleUse: invite.IsSingleUse,
		IsExpired:   isExpired,
		IsClaimed:   isClaimed,
	}
}

func (m *checklistInviteDtoMapper) ToDTOArray(invites []domain.ChecklistInvite, baseUrl string) []InviteResponse {
	responses := make([]InviteResponse, 0, len(invites))
	for _, invite := range invites {
		responses = append(responses, m.ToDTO(invite, baseUrl))
	}
	return responses
}
