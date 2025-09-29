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

func (controller *checklistController) DeleteChecklistById(_ context.Context, request DeleteChecklistByIdRequestObject) (DeleteChecklistByIdResponseObject, error) {
	if err := controller.service.DeleteChecklistById(request.ChecklistId); err == nil {
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

func (controller *checklistController) UpdateChecklistById(_ context.Context, request UpdateChecklistByIdRequestObject) (UpdateChecklistByIdResponseObject, error) {
	domainObject := controller.mapper.ToDomain(*request.Body)
	domainObject.Id = request.ChecklistId
	if checklist, err := controller.service.UpdateChecklist(domainObject); err == nil {
		dto := controller.mapper.ToDTO(checklist)
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

func (controller *checklistController) GetAllChecklists(_ context.Context, _ GetAllChecklistsRequestObject) (GetAllChecklistsResponseObject, error) {
	if checklists, err := controller.service.FindAllChecklists(); err == nil {
		dto := controller.mapper.ToDtoArray(checklists)
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

func (controller *checklistController) CreateChecklist(_ context.Context, request CreateChecklistRequestObject) (CreateChecklistResponseObject, error) {
	domainObject := controller.mapper.ToDomain(*request.Body)
	if checklist, err := controller.service.SaveChecklist(domainObject); err == nil {
		dto := controller.mapper.ToDTO(checklist)
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

func (controller *checklistController) GetChecklistById(_ context.Context, request GetChecklistByIdRequestObject) (GetChecklistByIdResponseObject, error) {
	if checklist, err := controller.service.FindChecklistById(request.ChecklistId); err == nil && checklist != nil {
		dto := controller.mapper.ToDTO(*checklist)
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

func NewChecklistController(service service.IChecklistService) IChecklistController {
	return &checklistController{
		service: service,
		mapper:  NewChecklistDtoMapper(),
	}
}
