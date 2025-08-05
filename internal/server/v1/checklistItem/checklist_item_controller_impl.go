package checklistItem

import (
	"context"
	"net/http"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/service"
)

type IChecklistItemController = StrictServerInterface

type checklistItemController struct {
	service service.IChecklistItemsService
	mapper  IChecklistItemDtoMapper
}

func (controller *checklistItemController) ListChecklistItems(_ context.Context, request ListChecklistItemsRequestObject) (ListChecklistItemsResponseObject, error) {
	sortOrder, err := domain.NewSortOrder((*string)(request.Params.Sort))
	if err != nil {
		return ListChecklistItems400JSONResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}, nil
	}

	if checklistItems, err := controller.service.FindAllChecklistItems(request.ChecklistId, request.Params.Completed, sortOrder); err == nil {
		dto := controller.mapper.MapDomainListToDtoList(checklistItems)
		return ListChecklistItems200JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return ListChecklistItems400JSONResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}, nil
	} else {
		return ListChecklistItems500JSONResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistItemController) DeleteChecklistItem(_ context.Context, request DeleteChecklistItemRequestObject) (DeleteChecklistItemResponseObject, error) {
	if err := controller.service.DeleteChecklistItemById(request.ChecklistId, request.ItemId); err == nil {
		return DeleteChecklistItem204JSONResponse{}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return DeleteChecklistItem404JSONResponse{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}, nil
	} else {
		return DeleteChecklistItem500JSONResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}, nil
	}
}

func (c *checklistItemController) ChangeChecklistItemOrder(_ context.Context, request ChangeChecklistItemOrderRequestObject) (ChangeChecklistItemOrderResponseObject, error) {
	sortOrder, err := domain.NewSortOrder((*string)(request.Params.SortOrder))
	if err != nil {
		return ChangeChecklistItemOrder400JSONResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}, nil
	}

	changeOrderRequest := domain.ChangeOrderRequest{
		NewOrderNumber:  request.Body.NewOrderNumber,
		ChecklistId:     request.ChecklistId,
		ChecklistItemId: request.ItemId,
		SortOrder:       sortOrder,
	}

	if response, err := c.service.ChangeChecklistItemOrder(changeOrderRequest); err == nil {
		return ChangeChecklistItemOrder200JSONResponse{
			NewOrderNumber: &response.OrderNumber,
			OldOrderNumber: nil,
		}, nil
	} else {
		switch err.ResponseCode() {
		case http.StatusBadRequest:
			return ChangeChecklistItemOrder400JSONResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}, nil
		case http.StatusNotFound:
			return ChangeChecklistItemOrder404JSONResponse{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}, nil
		default:
			return ChangeChecklistItemOrder500JSONResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}, nil
		}
	}
}

func (c *checklistItemController) CreateChecklistItem(_ context.Context, request CreateChecklistItemRequestObject) (CreateChecklistItemResponseObject, error) {
	domainObject := c.mapper.MapCreateRequestToDomain(*request.Body)
	if checklistItems, err := c.service.SaveChecklistItem(request.ChecklistId, domainObject); err == nil {
		dto := c.mapper.MapDomainToDto(checklistItems)
		return CreateChecklistItem201JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return CreateChecklistItem400JSONResponse{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}, nil
	} else {
		return CreateChecklistItem500JSONResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}, nil
	}
}

func (c *checklistItemController) CreateChecklistItemRow(_ context.Context, request CreateChecklistItemRowRequestObject) (CreateChecklistItemRowResponseObject, error) {
	domainRow := c.mapper.MapCreateChecklistItemRowRequestToDomain(*request.Body)
	if row, err := c.service.SaveChecklistItemRow(request.ChecklistId, request.ItemId, domainRow); err == nil {
		dto := c.mapper.MapChecklistItemRowDomainToDto(row)
		return CreateChecklistItemRow201JSONResponse(dto), nil
	} else {
		switch err.ResponseCode() {
		case http.StatusBadRequest:
			return CreateChecklistItemRow400JSONResponse{Code: http.StatusBadRequest, Message: err.Error()}, nil
		case http.StatusNotFound:
			return CreateChecklistItemRow404JSONResponse{Code: http.StatusNotFound, Message: err.Error()}, nil
		default:
			return CreateChecklistItemRow500JSONResponse{Code: http.StatusInternalServerError, Message: err.Error()}, nil
		}
	}
}

func (c *checklistItemController) GetChecklistItem(_ context.Context, request GetChecklistItemRequestObject) (GetChecklistItemResponseObject, error) {
	if checklistItem, err := c.service.FindChecklistItemById(request.ChecklistId, request.ItemId); err != nil {
		return GetChecklistItem500JSONResponse{Code: http.StatusInternalServerError, Message: err.Error()}, nil
	} else if checklistItem == nil {
		return GetChecklistItem404JSONResponse{
			Code:    http.StatusNotFound,
			Message: "Checklist item not found",
		}, nil
	} else {
		dto := c.mapper.MapDomainToDto(*checklistItem)
		return GetChecklistItem200JSONResponse(dto), nil
	}
}

func (c *checklistItemController) UpdateChecklistItem(_ context.Context, request UpdateChecklistItemRequestObject) (UpdateChecklistItemResponseObject, error) {
	domainObject := c.mapper.MapUpdateRequestToDomain(*request.Body)
	domainObject.Id = request.ItemId
	if updatedItem, err := c.service.UpdateChecklistItem(request.ChecklistId, domainObject); err != nil && err.ResponseCode() == 404 {
		return UpdateChecklistItem404JSONResponse{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}, nil
	} else if err != nil && err.ResponseCode() == 500 {
		return UpdateChecklistItem500JSONResponse{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}, nil
	} else {
		dto := c.mapper.MapDomainToDto(updatedItem)
		return UpdateChecklistItem200JSONResponse(dto), nil
	}
}

func NewChecklistItemController(
	service service.IChecklistItemsService) IChecklistItemController {
	return &checklistItemController{
		service: service,
		mapper:  NewChecklistItemMapper(),
	}
}
