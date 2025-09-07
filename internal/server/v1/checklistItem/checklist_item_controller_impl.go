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

func (controller *checklistItemController) ToggleChecklistItemComplete(_ context.Context, request ToggleChecklistItemCompleteRequestObject) (ToggleChecklistItemCompleteResponseObject, error) {
	// Directly mark the item as completed/uncompleted in a single atomic operation
	updatedItem, err := controller.service.ToggleCompleted(request.ChecklistId, request.ItemId, request.Body.Completed)
	if err == nil {
		dto := controller.mapper.MapDomainToDto(updatedItem)
		return ToggleChecklistItemComplete200JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return ToggleChecklistItemComplete404JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return ToggleChecklistItemComplete400JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return ToggleChecklistItemComplete500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *checklistItemController) GetAllChecklistItems(_ context.Context, request GetAllChecklistItemsRequestObject) (GetAllChecklistItemsResponseObject, error) {
	sortOrder, err := domain.NewSortOrder((*string)(request.Params.Sort))
	if err != nil {
		return GetAllChecklistItems400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	if checklistItems, err := controller.service.FindAllChecklistItems(request.ChecklistId, request.Params.Completed, sortOrder); err == nil {
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

func (controller *checklistItemController) DeleteChecklistItemRow(_ context.Context, request DeleteChecklistItemRowRequestObject) (DeleteChecklistItemRowResponseObject, error) {
	if err := controller.service.DeleteChecklistItemRow(request.ChecklistId, request.ItemId, request.RowId); err == nil {
		return DeleteChecklistItemRow204Response{}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return DeleteChecklistItemRow404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return DeleteChecklistItemRow500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (c *checklistItemController) ChangeChecklistItemOrderNumber(_ context.Context, request ChangeChecklistItemOrderNumberRequestObject) (ChangeChecklistItemOrderNumberResponseObject, error) {
	sortOrder, err := domain.NewSortOrder((*string)(request.Params.SortOrder))
	if err != nil {
		return ChangeChecklistItemOrderNumber400JSONResponse{
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
		return ChangeChecklistItemOrderNumber200JSONResponse{
			NewOrderNumber: &response.OrderNumber,
			OldOrderNumber: nil,
		}, nil
	} else {
		switch err.ResponseCode() {
		case http.StatusBadRequest:
			return ChangeChecklistItemOrderNumber400JSONResponse{
				Message: err.Error(),
			}, nil
		case http.StatusNotFound:
			return ChangeChecklistItemOrderNumber404JSONResponse{
				Message: err.Error(),
			}, nil
		default:
			return ChangeChecklistItemOrderNumber500JSONResponse{
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
			Message: err.Error(),
		}, nil
	} else {
		return CreateChecklistItem500JSONResponse{
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
			return CreateChecklistItemRow400JSONResponse{Message: err.Error()}, nil
		case http.StatusNotFound:
			return CreateChecklistItemRow404JSONResponse{Message: err.Error()}, nil
		default:
			return CreateChecklistItemRow500JSONResponse{Message: err.Error()}, nil
		}
	}
}

func (c *checklistItemController) GetChecklistItemBychecklistIdAndItemId(_ context.Context, request GetChecklistItemBychecklistIdAndItemIdRequestObject) (GetChecklistItemBychecklistIdAndItemIdResponseObject, error) {
	if checklistItem, err := c.service.FindChecklistItemById(request.ChecklistId, request.ItemId); err != nil {
		return GetChecklistItemBychecklistIdAndItemId500JSONResponse{Message: err.Error()}, nil
	} else if checklistItem == nil {
		return GetChecklistItemBychecklistIdAndItemId404JSONResponse{
			Message: "Checklist item not found",
		}, nil
	} else {
		dto := c.mapper.MapDomainToDto(*checklistItem)
		return GetChecklistItemBychecklistIdAndItemId200JSONResponse(dto), nil
	}
}

func (c *checklistItemController) UpdateChecklistItemBychecklistIdAndItemId(_ context.Context, request UpdateChecklistItemBychecklistIdAndItemIdRequestObject) (UpdateChecklistItemBychecklistIdAndItemIdResponseObject, error) {
	domainObject := c.mapper.MapUpdateRequestToDomain(*request.Body)
	domainObject.Id = request.ItemId
	if updatedItem, err := c.service.UpdateChecklistItem(request.ChecklistId, domainObject); err != nil && err.ResponseCode() == 404 {
		return UpdateChecklistItemBychecklistIdAndItemId404JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err != nil && err.ResponseCode() == 500 {
		return UpdateChecklistItemBychecklistIdAndItemId500JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		dto := c.mapper.MapDomainToDto(updatedItem)
		return UpdateChecklistItemBychecklistIdAndItemId200JSONResponse(dto), nil
	}
}

func NewChecklistItemController(
	service service.IChecklistItemsService) IChecklistItemController {
	return &checklistItemController{
		service: service,
		mapper:  NewChecklistItemMapper(),
	}
}
