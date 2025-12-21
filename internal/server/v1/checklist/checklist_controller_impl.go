package checklist

import (
	"context"
	"net/http"

	"com.raunlo.checklist/internal/core/service"
	serverutils "com.raunlo.checklist/internal/server/server_utils"
)

type IChecklistController = StrictServerInterface

type checklistController struct {
	service       service.IChecklistService
	inviteService service.IChecklistInviteService
	mapper        IChecklistDtoMapper
	inviteMapper  IChecklistInviteDtoMapper
	baseUrl       string
}

func (controller *checklistController) DeleteChecklistById(ctx context.Context, request DeleteChecklistByIdRequestObject) (DeleteChecklistByIdResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	if err := controller.service.DeleteChecklistById(domainContext, request.ChecklistId); err == nil {
		return DeleteChecklistById204JSONResponse{}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return DeleteChecklistById404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return DeleteChecklistById500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistController) UpdateChecklistById(ctx context.Context, request UpdateChecklistByIdRequestObject) (UpdateChecklistByIdResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	domainObject := controller.mapper.ToDomain(*request.Body)
	domainObject.Id = request.ChecklistId
	if checklist, err := controller.service.UpdateChecklist(domainContext, domainObject); err == nil {
		dto := controller.mapper.ToDTO(checklist, domainContext)
		return UpdateChecklistById200JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return UpdateChecklistById400JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return UpdateChecklistById404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return UpdateChecklistById500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistController) GetAllChecklists(ctx context.Context, _ GetAllChecklistsRequestObject) (GetAllChecklistsResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	if checklists, err := controller.service.FindAllChecklists(domainContext); err == nil {
		dto := controller.mapper.ToDtoArray(checklists, domainContext)
		return GetAllChecklists200JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return GetAllChecklists400JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return GetAllChecklists500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistController) CreateChecklist(ctx context.Context, request CreateChecklistRequestObject) (CreateChecklistResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	domainObject := controller.mapper.ToDomain(*request.Body)
	if checklist, err := controller.service.SaveChecklist(domainContext, domainObject); err == nil {
		dto := controller.mapper.ToDTO(checklist, domainContext)
		return CreateChecklist201JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return CreateChecklist400JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return CreateChecklist500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistController) GetChecklistById(ctx context.Context, request GetChecklistByIdRequestObject) (GetChecklistByIdResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	if checklist, err := controller.service.FindChecklistById(domainContext, request.ChecklistId); err == nil && checklist != nil {
		dto := controller.mapper.ToDTO(*checklist, domainContext)
		return GetChecklistById200JSONResponse(dto), nil
	} else if err == nil && checklist == nil {
		return GetChecklistById404JSONResponse{
			Message: "Checklist not found",
		}, nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return GetChecklistById400JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return GetChecklistById500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

// Invite methods

func (controller *checklistController) CreateChecklistInvite(ctx context.Context, request CreateChecklistInviteRequestObject) (CreateChecklistInviteResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	// Extract request parameters
	var expiresInHours *int
	if request.Body.ExpiresInHours != nil {
		expiresInHours = request.Body.ExpiresInHours
	}

	isSingleUse := request.Body.IsSingleUse

	// Call service
	invite, err := controller.inviteService.CreateInvite(domainContext, request.ChecklistId, expiresInHours, isSingleUse)
	if err == nil {
		dto := controller.inviteMapper.ToDTO(invite, controller.baseUrl)
		return CreateChecklistInvite201JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusForbidden {
		return CreateChecklistInvite403JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return CreateChecklistInvite404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return CreateChecklistInvite500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistController) GetChecklistInvites(ctx context.Context, request GetChecklistInvitesRequestObject) (GetChecklistInvitesResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	invites, err := controller.inviteService.GetActiveInvites(domainContext, request.ChecklistId)
	if err == nil {
		dtos := controller.inviteMapper.ToDTOArray(invites, controller.baseUrl)
		return GetChecklistInvites200JSONResponse(dtos), nil
	} else if err.ResponseCode() == http.StatusForbidden {
		return GetChecklistInvites403JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return GetChecklistInvites404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return GetChecklistInvites500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistController) RevokeChecklistInvite(ctx context.Context, request RevokeChecklistInviteRequestObject) (RevokeChecklistInviteResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	err := controller.inviteService.RevokeInvite(domainContext, request.InviteId)
	if err == nil {
		return RevokeChecklistInvite204Response{}, nil
	} else if err.ResponseCode() == http.StatusForbidden {
		return RevokeChecklistInvite403JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return RevokeChecklistInvite404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return RevokeChecklistInvite500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistController) ClaimInvite(ctx context.Context, request ClaimInviteRequestObject) (ClaimInviteResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	checklistId, err := controller.inviteService.ClaimInvite(domainContext, request.Token)
	if err == nil {
		return ClaimInvite200JSONResponse{
			ChecklistId: checklistId,
			Message:     "Successfully joined checklist",
		}, nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return ClaimInvite400JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return ClaimInvite404JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusUnauthorized {
		return ClaimInvite401JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return ClaimInvite500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func NewChecklistController(service service.IChecklistService, inviteService service.IChecklistInviteService, baseUrl string) IChecklistController {
	return &checklistController{
		service:       service,
		inviteService: inviteService,
		mapper:        NewChecklistDtoMapper(),
		inviteMapper:  NewChecklistInviteDtoMapper(),
		baseUrl:       baseUrl,
	}
}
