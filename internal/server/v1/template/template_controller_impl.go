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
