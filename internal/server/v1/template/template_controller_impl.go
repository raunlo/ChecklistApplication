package template

import (
	"context"
	"net/http"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/service"
	serverutils "com.raunlo.checklist/internal/server/server_utils"
)

type ITemplateController = StrictServerInterface

type templateController struct {
	service service.ITemplateService
	mapper  ITemplateDtoMapper
}

func NewTemplateController(
	service service.ITemplateService,
	mapper ITemplateDtoMapper,
) ITemplateController {
	return &templateController{
		service: service,
		mapper:  mapper,
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

	// Initialize empty items array if not provided
	if domainObject.Items == nil {
		domainObject.Items = make([]domain.TemplateItem, 0)
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
	domainObject.ID = request.TemplateId

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

func (controller *templateController) GetTemplatePreview(ctx context.Context, request GetTemplatePreviewRequestObject) (GetTemplatePreviewResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)
	existingItems, newItems, err := controller.service.GetTemplatePreview(domainContext, request.ChecklistId, request.TemplateId)

	if err == nil {
		dto := controller.mapper.ToTemplatePreviewDTO(existingItems, newItems)
		return GetTemplatePreview200JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return GetTemplatePreview404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return GetTemplatePreview500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *templateController) ApplyTemplate(ctx context.Context, request ApplyTemplateRequestObject) (ApplyTemplateResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	// Initialize selected items array if not provided
	selectedItemIds := make([]uint, 0)
	if request.Body != nil && request.Body.ItemIds != nil {
		selectedItemIds = request.Body.ItemIds
	}

	items, err := controller.service.ApplyTemplateToChecklist(domainContext, request.ChecklistId, request.TemplateId, selectedItemIds)

	if err == nil {
		// Convert domain ChecklistItems to ChecklistItemResponse DTOs
		response := make([]ChecklistItemResponse, 0)
		for _, item := range items {
			dto := ChecklistItemResponse{
				Id:          item.Id,
				Name:        item.Name,
				Completed:   item.Completed,
				OrderNumber: item.OrderNumber,
			}
			response = append(response, dto)
		}
		return ApplyTemplate200JSONResponse(response), nil
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

func (controller *templateController) CreateTemplateFromItems(ctx context.Context, request CreateTemplateFromItemsRequestObject) (CreateTemplateFromItemsResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	template, err := controller.service.CreateTemplateFromItems(
		domainContext,
		request.Body.ChecklistId,
		request.Body.Name,
		request.Body.Description,
		request.Body.ChecklistItemIds,
	)

	if err == nil {
		dto := controller.mapper.ToDTO(template)
		return CreateTemplateFromItems201JSONResponse(dto), nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return CreateTemplateFromItems400JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return CreateTemplateFromItems404JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return CreateTemplateFromItems500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}

func (controller *templateController) CreateChecklistFromTemplate(ctx context.Context, request CreateChecklistFromTemplateRequestObject) (CreateChecklistFromTemplateResponseObject, error) {
	domainContext := serverutils.CreateContext(ctx)

	checklist, err := controller.service.CreateChecklistFromTemplate(domainContext, request.TemplateId, request.Body.Name)

	if err == nil {
		response := ChecklistResponse{
			Id:       checklist.Id,
			Name:     checklist.Name,
			IsOwner:  true,
			IsShared: false,
			Owner:    checklist.Owner,
		}
		return CreateChecklistFromTemplate201JSONResponse(response), nil
	} else if err.ResponseCode() == http.StatusNotFound {
		return CreateChecklistFromTemplate404JSONResponse{
			Message: err.Error(),
		}, nil
	} else if err.ResponseCode() == http.StatusBadRequest {
		return CreateChecklistFromTemplate400JSONResponse{
			Message: err.Error(),
		}, nil
	} else {
		return CreateChecklistFromTemplate500JSONResponse{
			Message: err.Error(),
		}, nil
	}
}
