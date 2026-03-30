package template

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

type ITemplateController = StrictServerInterface

type templateController struct {
	service       service.ITemplateService
	inviteService service.ITemplateInviteService
	mapper        ITemplateDtoMapper
	baseUrl       serverAuth.BaseUrl
}

func NewTemplateController(
	service service.ITemplateService,
	inviteService service.ITemplateInviteService,
	mapper ITemplateDtoMapper,
	baseUrl serverAuth.BaseUrl,
) ITemplateController {
	return &templateController{
		service:       service,
		inviteService: inviteService,
		mapper:        mapper,
		baseUrl:       baseUrl,
	}
}

func (controller *templateController) GetAllTemplates(ctx context.Context, _ GetAllTemplatesRequestObject) (GetAllTemplatesResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	templates, err := controller.service.FindAllTemplates(domainContext)

	if err == nil {
		response := controller.mapper.ToTemplateDtoArray(templates)
		return GetAllTemplates200JSONResponse(response), nil
	} else {
		return GetAllTemplates500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *templateController) CreateTemplate(ctx context.Context, request CreateTemplateRequestObject) (CreateTemplateResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	domainObject := controller.mapper.ToDomain(*request.Body)

	// Initialize empty rows array if not provided
	if domainObject.Rows == nil {
		domainObject.Rows = make([]domain.TemplateRow, 0)
	}

	if template, err := controller.service.SaveTemplate(domainContext, domainObject); err == nil {
		dto := controller.mapper.ToDTO(template)
		return CreateTemplate201JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return CreateTemplate400JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return CreateTemplate500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *templateController) GetTemplateById(ctx context.Context, request GetTemplateByIdRequestObject) (GetTemplateByIdResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	if template, err := controller.service.FindTemplateById(domainContext, request.TemplateId); err == nil && template != nil {
		dto := controller.mapper.ToDTO(*template)
		return GetTemplateById200JSONResponse(dto), nil
	} else if err == nil && template == nil {
		return GetTemplateById404JSONResponse{
			Message: "Template not found",
		}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return GetTemplateById404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return GetTemplateById500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *templateController) UpdateTemplate(ctx context.Context, request UpdateTemplateRequestObject) (UpdateTemplateResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	domainObject := controller.mapper.ToDomain(*request.Body)
	domainObject.Id = request.TemplateId

	if template, err := controller.service.UpdateTemplate(domainContext, domainObject); err == nil {
		dto := controller.mapper.ToDTO(template)
		return UpdateTemplate200JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return UpdateTemplate400JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return UpdateTemplate404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return UpdateTemplate500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *templateController) DeleteTemplate(ctx context.Context, request DeleteTemplateRequestObject) (DeleteTemplateResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	if err := controller.service.DeleteTemplate(domainContext, request.TemplateId); err == nil {
		return DeleteTemplate204Response{}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return DeleteTemplate404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return DeleteTemplate500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *templateController) ApplyTemplate(ctx context.Context, request ApplyTemplateRequestObject) (ApplyTemplateResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	item, err := controller.service.ApplyTemplateToChecklist(domainContext, request.ChecklistId, request.TemplateId)

	if err == nil {
		dto := ChecklistItemResponse{
			Id:          item.Id,
			Name:        item.Name,
			Completed:   item.Completed,
			OrderNumber: item.OrderNumber,
		}
		return ApplyTemplate200JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return ApplyTemplate404JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return ApplyTemplate400JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return ApplyTemplate500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *templateController) CreateTemplateFromItem(ctx context.Context, request CreateTemplateFromItemRequestObject) (CreateTemplateFromItemResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	template, err := controller.service.CreateTemplateFromItem(
		domainContext,
		request.Body.ChecklistId,
		request.Body.Name,
		request.Body.Description,
		request.Body.ChecklistItemId,
	)

	if err == nil {
		dto := controller.mapper.ToDTO(template)
		return CreateTemplateFromItem201JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return CreateTemplateFromItem400JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return CreateTemplateFromItem404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return CreateTemplateFromItem500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

// Invite methods

func (controller *templateController) toInviteDTO(invite domain.TemplateInvite) TemplateInviteResponse {
	isExpired := invite.ExpiresAt != nil && invite.ExpiresAt.Before(time.Now())
	isClaimed := invite.ClaimedAt != nil
	inviteUrl := string(controller.baseUrl) + "/template-invites/" + invite.InviteToken + "/claim"

	return TemplateInviteResponse{
		Id:          invite.Id,
		TemplateId:  invite.TemplateId,
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

func (controller *templateController) GetTemplateInvites(ctx context.Context, request GetTemplateInvitesRequestObject) (GetTemplateInvitesResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	invites, err := controller.inviteService.GetActiveInvites(domainContext, request.TemplateId)
	if err == nil {
		dtos := make([]TemplateInviteResponse, 0, len(invites))
		for _, inv := range invites {
			dtos = append(dtos, controller.toInviteDTO(inv))
		}
		return GetTemplateInvites200JSONResponse(dtos), nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return GetTemplateInvites404JSONResponse{Message: err.Error()}, nil
	} else {
		log.Printf("Error getting template invites: %v", err)
		return GetTemplateInvites500JSONResponse{Message: "Failed to retrieve invites"}, nil
	}
}

func (controller *templateController) CreateTemplateInvite(ctx context.Context, request CreateTemplateInviteRequestObject) (CreateTemplateInviteResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	if request.Body == nil {
		return CreateTemplateInvite500JSONResponse{Message: "Invalid request body"}, nil
	}

	var name *string
	if request.Body.Name != nil && *request.Body.Name != "" {
		name = request.Body.Name
	}

	var expiresInHours *int
	if request.Body.ExpiresInHours != nil {
		if *request.Body.ExpiresInHours < 1 || *request.Body.ExpiresInHours > 8760 {
			return CreateTemplateInvite500JSONResponse{Message: "Expiration hours must be between 1 and 8760 (1 year)"}, nil
		}
		expiresInHours = request.Body.ExpiresInHours
	}

	invite, err := controller.inviteService.CreateInvite(domainContext, request.TemplateId, name, expiresInHours, request.Body.IsSingleUse)
	if err == nil {
		dto := controller.toInviteDTO(invite)
		return CreateTemplateInvite201JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return CreateTemplateInvite404JSONResponse{Message: "Template not found"}, nil
	} else {
		log.Printf("Error creating template invite: %v", err)
		return CreateTemplateInvite500JSONResponse{Message: "Failed to create invite"}, nil
	}
}

func (controller *templateController) RevokeTemplateInvite(ctx context.Context, request RevokeTemplateInviteRequestObject) (RevokeTemplateInviteResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	err := controller.inviteService.RevokeInvite(domainContext, request.InviteId)
	if err == nil {
		return RevokeTemplateInvite204Response{}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return RevokeTemplateInvite404JSONResponse{Message: "Invite not found"}, nil
	} else {
		log.Printf("Error revoking template invite: %v", err)
		return RevokeTemplateInvite500JSONResponse{Message: "Failed to revoke invite"}, nil
	}
}

func (controller *templateController) ClaimTemplateInvite(ctx context.Context, request ClaimTemplateInviteRequestObject) (ClaimTemplateInviteResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	templateId, err := controller.inviteService.ClaimInvite(domainContext, request.Token)
	if err == nil {
		msg := "Successfully joined template"
		return ClaimTemplateInvite200JSONResponse{
			TemplateId: templateId,
			Message:    &msg,
		}, nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return ClaimTemplateInvite400JSONResponse{Message: err.Error()}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return ClaimTemplateInvite404JSONResponse{Message: "Invite not found"}, nil
	} else if err.ResponseCode() == http.StatusUnauthorized {
		return ClaimTemplateInvite401JSONResponse{Message: "Authentication required"}, nil
	} else {
		log.Printf("Error claiming template invite: %v", err)
		return ClaimTemplateInvite500JSONResponse{Message: "Failed to claim invite"}, nil
	}
}

func (controller *templateController) LeaveSharedTemplate(ctx context.Context, request LeaveSharedTemplateRequestObject) (LeaveSharedTemplateResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	err := controller.service.LeaveSharedTemplate(domainContext, request.TemplateId)
	if err == nil {
		return LeaveSharedTemplate204Response{}, nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return LeaveSharedTemplate400JSONResponse{Message: err.Error()}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return LeaveSharedTemplate404JSONResponse{Message: err.Error()}, nil
	} else if err.ResponseCode() == http.StatusUnauthorized {
		return LeaveSharedTemplate401JSONResponse{Message: err.Error()}, nil
	} else {
		log.Printf("Error leaving shared template: %v", err)
		return LeaveSharedTemplate500JSONResponse{Message: "Failed to leave template"}, nil
	}
}
