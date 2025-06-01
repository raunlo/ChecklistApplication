package checklistItem

import (
	"context"
	"net/http"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/service"
	"com.raunlo.checklist/internal/util"
)

type IChecklistItemController = StrictServerInterface

type checklistItemController struct {
	service service.IChecklistItemsService
	mapper  IChecklistItemDtoMapper
}

func (controller *checklistItemController) GetAllChecklistItems(_ context.Context, request GetAllChecklistItemsRequestObject) (GetAllChecklistItemsResponseObject, error) {
	var sort domain.SortOrder
	if s := request.Params.Sort; s != nil {
		sort = domain.NewSortOrder(util.StrPointer(string(*s)))
	}

	if checklistItems, err := controller.service.FindAllChecklistItems(request.ChecklistId, request.Params.Completed, &sort); err == nil {
		dto := controller.mapper.MapDomainListToDtoList(checklistItems)
		return GetAllChecklistItems200JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return GetAllChecklistItems400JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return GetAllChecklistItems500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistItemController) DeleteChecklistItemById(_ context.Context, request DeleteChecklistItemByIdRequestObject) (DeleteChecklistItemByIdResponseObject, error) {
	if err := controller.service.DeleteChecklistItemById(request.ChecklistId, request.ItemId); err == nil {
		return DeleteChecklistItemById204JSONResponse{}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return DeleteChecklistItemById404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return DeleteChecklistItemById500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistItemController) ChangeChecklistItemOrderNumber(_ context.Context, request ChangeChecklistItemOrderNumberRequestObject) (ChangeChecklistItemOrderNumberResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (controller *checklistItemController) CreateChecklistItem(_ context.Context, request CreateChecklistItemRequestObject) (CreateChecklistItemResponseObject, error) {
	domainObject := controller.mapper.MapDtoToDomain(*request.Body)
	if checklistItems, err := controller.service.SaveChecklistItem(request.ChecklistId, domainObject); err == nil {
		dto := controller.mapper.MapDomainToDto(checklistItems)
		return CreateChecklistItem201JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return CreateChecklistItem400JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return CreateChecklistItem500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func NewChecklistItemController(
	service service.IChecklistItemsService) IChecklistItemController {
	return &checklistItemController{
		service: service,
		mapper:  NewChecklistItemMapper(),
	}
}
