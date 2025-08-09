package checklist

import (
	"context"
	"net/http"

	"com.raunlo.checklist/internal/core/service"
)

type IChecklistController = StrictServerInterface

type checklistController struct {
	service service.IChecklistService
	mapper  IChecklistDtoMapper
}

func (controller *checklistController) DeleteChecklist(_ context.Context, request DeleteChecklistRequestObject) (DeleteChecklistResponseObject, error) {
	if err := controller.service.DeleteChecklistById(request.ChecklistId); err == nil {
		return DeleteChecklist204JSONResponse{}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return DeleteChecklist404JSONResponse{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}, nil
	} else {
		return DeleteChecklist500JSONResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistController) UpdateChecklist(_ context.Context, request UpdateChecklistRequestObject) (UpdateChecklistResponseObject, error) {
	domainObject := controller.mapper.ToDomain(*request.Body)
	domainObject.Id = request.ChecklistId
	if checklist, err := controller.service.UpdateChecklist(domainObject); err == nil {
		dto := controller.mapper.ToDTO(checklist)
		return UpdateChecklist200JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return UpdateChecklist400JSONResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return UpdateChecklist404JSONResponse{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}, nil
	} else {
		return UpdateChecklist500JSONResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistController) ListChecklists(_ context.Context, _ ListChecklistsRequestObject) (ListChecklistsResponseObject, error) {

	if checklists, err := controller.service.FindAllChecklists(); err == nil {
		dto := controller.mapper.ToDtoArray(checklists)
		return ListChecklists200JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return ListChecklists400JSONResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}, nil
	} else {
		return ListChecklists500JSONResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistController) CreateChecklist(_ context.Context, request CreateChecklistRequestObject) (CreateChecklistResponseObject, error) {
	domainObject := controller.mapper.ToDomain(*request.Body)
	if checklist, err := controller.service.SaveChecklist(domainObject); err == nil {
		dto := controller.mapper.ToDTO(checklist)
		return CreateChecklist201JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return CreateChecklist400JSONResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}, nil
	} else {
		return CreateChecklist500JSONResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistController) GetChecklist(_ context.Context, request GetChecklistRequestObject) (GetChecklistResponseObject, error) {
	if checklist, err := controller.service.FindChecklistById(request.ChecklistId); err == nil && checklist != nil {
		dto := controller.mapper.ToDTO(*checklist)
		return GetChecklist200JSONResponse(dto), nil
	} else if err == nil && checklist == nil {
		return GetChecklist404JSONResponse{
			Code:    http.StatusNotFound,
			Message: "Checklist not found",
		}, nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return GetChecklist400JSONResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}, nil
	} else {
		return GetChecklist500JSONResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}, nil
	}
}

func NewChecklistController(service service.IChecklistService) IChecklistController {
	return &checklistController{
		service: service,
		mapper:  NewChecklistDtoMapper(),
	}
}
