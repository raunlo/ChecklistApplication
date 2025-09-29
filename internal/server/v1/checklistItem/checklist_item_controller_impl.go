package checklistItem

import (
	"context"
	"net/http"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/service"
	serverutils "com.raunlo.checklist/internal/server/server_utils"
)

type IChecklistItemController = StrictServerInterface

type checklistItemController struct {
	service service.IChecklistItemsService
	mapper  IChecklistItemDtoMapper
}

func (controller *checklistItemController) ToggleChecklistItemComplete(ctx context.Context, request ToggleChecklistItemCompleteRequestObject) (ToggleChecklistItemCompleteResponseObject, error) {
	// Directly mark the item as completed/uncompleted in a single atomic operation
	domainContext := serverutils.CreateContext(ctx)
	updatedItem, err := controller.service.ToggleCompleted(domainContext, request.ChecklistId, request.ItemId, request.Body.Completed)
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

func (controller *checklistItemController) GetAllChecklistItems(ctx context.Context, request GetAllChecklistItemsRequestObject) (GetAllChecklistItemsResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	sortOrder, err := domain.NewSortOrder((*string)(request.Params.Sort))
	if err != nil {
		return GetAllChecklistItems400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	if checklistItems, err := controller.service.FindAllChecklistItems(domainContext, request.ChecklistId, request.Params.Completed, sortOrder); err == nil {
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

func (controller *checklistItemController) DeleteChecklistItemById(ctx context.Context, request DeleteChecklistItemByIdRequestObject) (DeleteChecklistItemByIdResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	if err := controller.service.DeleteChecklistItemById(domainContext, request.ChecklistId, request.ItemId); err == nil {
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

func (controller *checklistItemController) DeleteChecklistItemRow(ctx context.Context, request DeleteChecklistItemRowRequestObject) (DeleteChecklistItemRowResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	if err := controller.service.DeleteChecklistItemRow(domainContext, request.ChecklistId, request.ItemId, request.RowId); err == nil {
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

func (c *checklistItemController) ChangeChecklistItemOrderNumber(ctx context.Context, request ChangeChecklistItemOrderNumberRequestObject) (ChangeChecklistItemOrderNumberResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
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

	if response, err := c.service.ChangeChecklistItemOrder(domainContext, changeOrderRequest); err == nil {
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

func (c *checklistItemController) CreateChecklistItem(ctx context.Context, request CreateChecklistItemRequestObject) (CreateChecklistItemResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	domainObject := c.mapper.MapCreateRequestToDomain(*request.Body)
	if checklistItems, err := c.service.SaveChecklistItem(domainContext, request.ChecklistId, domainObject); err == nil {
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

func (c *checklistItemController) CreateChecklistItemRow(ctx context.Context, request CreateChecklistItemRowRequestObject) (CreateChecklistItemRowResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	domainRow := c.mapper.MapCreateChecklistItemRowRequestToDomain(*request.Body)
	if row, err := c.service.SaveChecklistItemRow(domainContext, request.ChecklistId, request.ItemId, domainRow); err == nil {
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

func (c *checklistItemController) GetChecklistItemBychecklistIdAndItemId(ctx context.Context, request GetChecklistItemBychecklistIdAndItemIdRequestObject) (GetChecklistItemBychecklistIdAndItemIdResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	if checklistItem, err := c.service.FindChecklistItemById(domainContext, request.ChecklistId, request.ItemId); err != nil {
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

func (c *checklistItemController) UpdateChecklistItemBychecklistIdAndItemId(ctx context.Context, request UpdateChecklistItemBychecklistIdAndItemIdRequestObject) (UpdateChecklistItemBychecklistIdAndItemIdResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	domainObject := c.mapper.MapUpdateRequestToDomain(*request.Body)
	domainObject.Id = request.ItemId
	if updatedItem, err := c.service.UpdateChecklistItem(domainContext, request.ChecklistId, domainObject); err != nil && err.ResponseCode() == 404 {
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
	service service.IChecklistItemsService,
) IChecklistItemController {
	return &checklistItemController{
		service: service,
		mapper:  NewChecklistItemMapper(),
	}
}
