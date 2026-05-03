package workspace

import (
	"context"
	"log"
	"net/http"
	"time"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/service"
	serverAuth "com.raunlo.checklist/internal/server/auth"
	serverutils "com.raunlo.checklist/internal/server/server_utils"
)

type IWorkspaceController = StrictServerInterface

type workspaceController struct {
	service       service.IWorkspaceService
	inviteService service.IWorkspaceInviteService
	baseUrl       serverAuth.BaseUrl
}

func NewWorkspaceController(
	service service.IWorkspaceService,
	inviteService service.IWorkspaceInviteService,
	baseUrl serverAuth.BaseUrl,
) IWorkspaceController {
	return &workspaceController{
		service:       service,
		inviteService: inviteService,
		baseUrl:       baseUrl,
	}
}

func toWorkspaceResponse(w domain.Workspace) WorkspaceResponse {
	return WorkspaceResponse{
		Id:          w.Id,
		Name:        w.Name,
		Description: w.Description,
		IsOwner:     w.IsOwner,
		IsDefault:   w.IsDefault,
		MemberCount: w.MemberCount,
	}
}

func (c *workspaceController) toInviteDTO(invite domain.WorkspaceInvite) WorkspaceInviteResponse {
	isExpired := invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now())
	isClaimed := invite.ClaimedAt != nil
	inviteUrl := string(c.baseUrl) + "/workspace-invites/" + invite.InviteToken + "/claim"

	return WorkspaceInviteResponse{
		Id:          invite.Id,
		WorkspaceId: invite.WorkspaceId,
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

func (c *workspaceController) GetAllWorkspaces(ctx context.Context, _ GetAllWorkspacesRequestObject) (GetAllWorkspacesResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	workspaces, err := c.service.FindAllWorkspaces(domainCtx)
	if err != nil {
		return GetAllWorkspaces500JSONResponse{ErrorResponseJSONResponse: ErrorResponseJSONResponse{Message: err.Error()}}, nil
	}
	dtos := make([]WorkspaceResponse, 0, len(workspaces))
	for _, w := range workspaces {
		dtos = append(dtos, toWorkspaceResponse(w))
	}
	return GetAllWorkspaces200JSONResponse(dtos), nil
}

func (c *workspaceController) CreateWorkspace(ctx context.Context, request CreateWorkspaceRequestObject) (CreateWorkspaceResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	w := domain.Workspace{
		Name:        request.Body.Name,
		Description: request.Body.Description,
	}
	created, err := c.service.CreateWorkspace(domainCtx, w)
	if err == nil {
		return CreateWorkspace201JSONResponse(toWorkspaceResponse(created)), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return CreateWorkspace400JSONResponse{ErrorResponseJSONResponse: ErrorResponseJSONResponse{Message: err.Error()}}, nil
	} else {
		return CreateWorkspace500JSONResponse{Message: err.Error()}, nil
	}
}

func (c *workspaceController) GetWorkspaceById(ctx context.Context, request GetWorkspaceByIdRequestObject) (GetWorkspaceByIdResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	w, err := c.service.FindWorkspaceById(domainCtx, request.WorkspaceId)
	if err == nil && w != nil {
		return GetWorkspaceById200JSONResponse(toWorkspaceResponse(*w)), nil
	} else if err == nil && w == nil {
		return GetWorkspaceById404JSONResponse{Message: "Workspace not found"}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return GetWorkspaceById404JSONResponse{Message: err.Error()}, nil
	} else {
		return GetWorkspaceById500JSONResponse{Message: err.Error()}, nil
	}
}

func (c *workspaceController) UpdateWorkspace(ctx context.Context, request UpdateWorkspaceRequestObject) (UpdateWorkspaceResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	w := domain.Workspace{
		Id:          request.WorkspaceId,
		Name:        request.Body.Name,
		Description: request.Body.Description,
	}
	updated, err := c.service.UpdateWorkspace(domainCtx, w)
	if err == nil {
		return UpdateWorkspace200JSONResponse(toWorkspaceResponse(updated)), nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return UpdateWorkspace404JSONResponse{Message: err.Error()}, nil
	} else {
		return UpdateWorkspace500JSONResponse{Message: err.Error()}, nil
	}
}

func (c *workspaceController) DeleteWorkspace(ctx context.Context, request DeleteWorkspaceRequestObject) (DeleteWorkspaceResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	if err := c.service.DeleteWorkspace(domainCtx, request.WorkspaceId); err == nil {
		return DeleteWorkspace204Response{}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return DeleteWorkspace404JSONResponse{Message: err.Error()}, nil
	} else {
		return DeleteWorkspace500JSONResponse{Message: err.Error()}, nil
	}
}

func (c *workspaceController) LeaveWorkspace(ctx context.Context, request LeaveWorkspaceRequestObject) (LeaveWorkspaceResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	if err := c.service.LeaveWorkspace(domainCtx, request.WorkspaceId); err == nil {
		return LeaveWorkspace204Response{}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return LeaveWorkspace404JSONResponse{Message: err.Error()}, nil
	} else {
		return LeaveWorkspace500JSONResponse{Message: err.Error()}, nil
	}
}

func (c *workspaceController) GetWorkspaceMembers(ctx context.Context, request GetWorkspaceMembersRequestObject) (GetWorkspaceMembersResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	members, err := c.service.GetMembers(domainCtx, request.WorkspaceId)
	if err == nil {
		dtos := make([]WorkspaceMemberResponse, 0, len(members))
		for _, m := range members {
			dtos = append(dtos, WorkspaceMemberResponse{
				MemberId: m.MemberId,
				Name:     m.Name,
				IsOwner:  m.IsOwner,
			})
		}
		return GetWorkspaceMembers200JSONResponse(dtos), nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return GetWorkspaceMembers404JSONResponse{Message: err.Error()}, nil
	} else {
		return GetWorkspaceMembers500JSONResponse{Message: err.Error()}, nil
	}
}

func (c *workspaceController) RemoveWorkspaceMember(ctx context.Context, request RemoveWorkspaceMemberRequestObject) (RemoveWorkspaceMemberResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	if err := c.service.RemoveMember(domainCtx, request.WorkspaceId, request.MemberId); err == nil {
		return RemoveWorkspaceMember204Response{}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return RemoveWorkspaceMember404JSONResponse{Message: err.Error()}, nil
	} else {
		return RemoveWorkspaceMember500JSONResponse{Message: err.Error()}, nil
	}
}

func (c *workspaceController) GetWorkspaceInvites(ctx context.Context, request GetWorkspaceInvitesRequestObject) (GetWorkspaceInvitesResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	invites, err := c.inviteService.GetActiveInvites(domainCtx, request.WorkspaceId)
	if err == nil {
		dtos := make([]WorkspaceInviteResponse, 0, len(invites))
		for _, inv := range invites {
			dtos = append(dtos, c.toInviteDTO(inv))
		}
		return GetWorkspaceInvites200JSONResponse(dtos), nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return GetWorkspaceInvites404JSONResponse{Message: err.Error()}, nil
	} else {
		log.Printf("Error getting workspace invites: %v", err)
		return GetWorkspaceInvites500JSONResponse{Message: "Failed to retrieve invites"}, nil
	}
}

func (c *workspaceController) CreateWorkspaceInvite(ctx context.Context, request CreateWorkspaceInviteRequestObject) (CreateWorkspaceInviteResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	if request.Body == nil {
		return CreateWorkspaceInvite400JSONResponse{ErrorResponseJSONResponse: ErrorResponseJSONResponse{Message: "Invalid request body"}}, nil
	}

	var expiresInHours *int
	if request.Body.ExpiresInHours != nil {
		if *request.Body.ExpiresInHours < 1 || *request.Body.ExpiresInHours > 8760 {
			return CreateWorkspaceInvite400JSONResponse{ErrorResponseJSONResponse: ErrorResponseJSONResponse{Message: "Expiration hours must be between 1 and 8760 (1 year)"}}, nil
		}
		expiresInHours = request.Body.ExpiresInHours
	}

	invite, err := c.inviteService.CreateInvite(domainCtx, request.WorkspaceId, request.Body.Name, expiresInHours, request.Body.IsSingleUse)
	if err == nil {
		return CreateWorkspaceInvite201JSONResponse(c.toInviteDTO(invite)), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return CreateWorkspaceInvite400JSONResponse{ErrorResponseJSONResponse: ErrorResponseJSONResponse{Message: err.Error()}}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return CreateWorkspaceInvite404JSONResponse{Message: "Workspace not found"}, nil
	} else {
		log.Printf("Error creating workspace invite: %v", err)
		return CreateWorkspaceInvite500JSONResponse{Message: err.Error()}, nil
	}
}

func (c *workspaceController) RevokeWorkspaceInvite(ctx context.Context, request RevokeWorkspaceInviteRequestObject) (RevokeWorkspaceInviteResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	if err := c.inviteService.RevokeInvite(domainCtx, request.WorkspaceId, request.InviteId); err == nil {
		return RevokeWorkspaceInvite204Response{}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return RevokeWorkspaceInvite404JSONResponse{Message: "Invite not found"}, nil
	} else {
		log.Printf("Error revoking workspace invite: %v", err)
		return RevokeWorkspaceInvite500JSONResponse{Message: "Failed to revoke invite"}, nil
	}
}

func (c *workspaceController) ClaimWorkspaceInvite(ctx context.Context, request ClaimWorkspaceInviteRequestObject) (ClaimWorkspaceInviteResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	workspaceId, err := c.inviteService.ClaimInvite(domainCtx, request.Token)
	if err == nil {
		msg := "Successfully joined workspace"
		return ClaimWorkspaceInvite200JSONResponse{
			WorkspaceId: workspaceId,
			Message:     &msg,
		}, nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return ClaimWorkspaceInvite400JSONResponse{ErrorResponseJSONResponse: ErrorResponseJSONResponse{Message: err.Error()}}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return ClaimWorkspaceInvite404JSONResponse{Message: "Invite not found"}, nil
	} else if err.ResponseCode() == http.StatusUnauthorized {
		return ClaimWorkspaceInvite401JSONResponse{Message: "Authentication required"}, nil
	} else {
		log.Printf("Error claiming workspace invite: %v", err)
		return ClaimWorkspaceInvite500JSONResponse{Message: "Failed to claim invite"}, nil
	}
}

func (c *workspaceController) GetWorkspaceTemplates(ctx context.Context, request GetWorkspaceTemplatesRequestObject) (GetWorkspaceTemplatesResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	templates, err := c.service.GetWorkspaceTemplates(domainCtx, request.WorkspaceId)
	if err == nil {
		dtos := make([]TemplateResponse, 0, len(templates))
		for _, t := range templates {
			rows := make([]TemplateRowResponse, 0, len(t.Rows))
			for _, r := range t.Rows {
				rows = append(rows, TemplateRowResponse{
					Id:         r.Id,
					TemplateId: r.TemplateId,
					Name:       r.Name,
					Position:   r.Position,
					CreatedAt:  r.CreatedAt,
					UpdatedAt:  r.UpdatedAt,
				})
			}
			wsIds := make([]uint, len(t.WorkspaceIds))
			copy(wsIds, t.WorkspaceIds)
			dtos = append(dtos, TemplateResponse{
				Id:           t.Id,
				Name:         t.Name,
				Description:  t.Description,
				UserId:       t.UserId,
				IsOwner:      t.IsOwner,
				Rows:         rows,
				WorkspaceIds: wsIds,
				CreatedAt:    t.CreatedAt,
				UpdatedAt:    t.UpdatedAt,
			})
		}
		return GetWorkspaceTemplates200JSONResponse(dtos), nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return GetWorkspaceTemplates404JSONResponse{Message: err.Error()}, nil
	} else {
		return GetWorkspaceTemplates500JSONResponse{Message: err.Error()}, nil
	}
}

func (c *workspaceController) GetWorkspaceChecklists(ctx context.Context, request GetWorkspaceChecklistsRequestObject) (GetWorkspaceChecklistsResponseObject, error) {
	domainCtx := serverutils.CreateContext(ctx)
	checklists, err := c.service.GetWorkspaceChecklists(domainCtx, request.WorkspaceId)
	if err == nil {
		currentUserId, _ := domain.GetUserIdFromContext(domainCtx)
		dtos := make([]ChecklistWithStats, 0, len(checklists))
		for _, cl := range checklists {
			isOwner := cl.Owner == currentUserId
			isShared := len(cl.SharedWith) > 0
			dto := ChecklistWithStats{
				Id:       cl.Id,
				Name:     cl.Name,
				IsOwner:  isOwner,
				IsShared: isShared,
				Stats: struct {
					CompletedItems uint `json:"completedItems"`
					TotalItems     uint `json:"totalItems"`
				}{
					CompletedItems: cl.Stats.CompletedItems,
					TotalItems:     cl.Stats.TotalItems,
				},
			}
			if isOwner {
				sharedCount := float32(len(cl.SharedWith))
				dto.NumberOfSharedUsers = &sharedCount
			}
			dtos = append(dtos, dto)
		}
		return GetWorkspaceChecklists200JSONResponse(dtos), nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return GetWorkspaceChecklists404JSONResponse{Message: err.Error()}, nil
	} else {
		return GetWorkspaceChecklists500JSONResponse{Message: err.Error()}, nil
	}
}
