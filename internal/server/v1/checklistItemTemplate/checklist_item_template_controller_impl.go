package checklistItemTemplate

import (
	"context"
	"net/http"

	"com.raunlo.checklist/internal/core/service"
)

type IChecklistItemTemplateController = StrictServerInterface

type checklistItemTemplateController struct {
	service service.IChecklistItemTemplateService
	mapper  IChecklistItemTemplateDtoMapper
}

func (c *checklistItemTemplateController) GetAllChecklistItemTemplates(_ context.Context, _ GetAllChecklistItemTemplatesRequestObject) (GetAllChecklistItemTemplatesResponseObject, error) {
	if templates, err := c.service.GetAllChecklistTemplates(); err != nil {
		return GetAllChecklistItemTemplates500JSONResponse{Message: err.Error()}, nil
	} else {
		dto := c.mapper.MapDomainListToDtoList(templates)
		return GetAllChecklistItemTemplates200JSONResponse(dto), nil
	}
}

func (c *checklistItemTemplateController) CreateChecklistItemTemplate(_ context.Context, request CreateChecklistItemTemplateRequestObject) (CreateChecklistItemTemplateResponseObject, error) {
	domainObject := c.mapper.MapCreateRequestToDomain(*request.Body)
	if template, err := c.service.SaveChecklistTemplate(domainObject); err != nil {
		if err.ResponseCode() == http.StatusBadRequest {
			return CreateChecklistItemTemplate400JSONResponse{Message: err.Error()}, nil
		}
		return CreateChecklistItemTemplate500JSONResponse{Message: err.Error()}, nil
	} else {
		dto := c.mapper.MapDomainToDto(template)
		return CreateChecklistItemTemplate201JSONResponse(dto), nil
	}
}

func (c *checklistItemTemplateController) GetChecklistItemTemplateById(_ context.Context, request GetChecklistItemTemplateByIdRequestObject) (GetChecklistItemTemplateByIdResponseObject, error) {
	if template, err := c.service.FindChecklistTemplateById(request.TemplateId); err != nil {
		return GetChecklistItemTemplateById500JSONResponse{Message: err.Error()}, nil
	} else if template == nil {
		return GetChecklistItemTemplateById404JSONResponse{Message: "Checklist item template not found"}, nil
	} else {
		dto := c.mapper.MapDomainToDto(*template)
		return GetChecklistItemTemplateById200JSONResponse(dto), nil
	}
}

func (c *checklistItemTemplateController) UpdateChecklistItemTemplateById(_ context.Context, request UpdateChecklistItemTemplateByIdRequestObject) (UpdateChecklistItemTemplateByIdResponseObject, error) {
	domainObject := c.mapper.MapUpdateRequestToDomain(*request.Body)
	domainObject.Id = request.TemplateId
	if template, err := c.service.UpdateChecklistTemplate(domainObject); err != nil {
		if err.ResponseCode() == http.StatusNotFound {
			return UpdateChecklistItemTemplateById404JSONResponse{Message: err.Error()}, nil
		}
		return UpdateChecklistItemTemplateById500JSONResponse{Message: err.Error()}, nil
	} else {
		dto := c.mapper.MapDomainToDto(template)
		return UpdateChecklistItemTemplateById200JSONResponse(dto), nil
	}
}

func (c *checklistItemTemplateController) DeleteChecklistItemTemplateById(_ context.Context, request DeleteChecklistItemTemplateByIdRequestObject) (DeleteChecklistItemTemplateByIdResponseObject, error) {
	if err := c.service.DeleteChecklistTemplateById(request.TemplateId); err != nil {
		if err.ResponseCode() == http.StatusNotFound {
			return DeleteChecklistItemTemplateById404JSONResponse{Message: err.Error()}, nil
		}
		return DeleteChecklistItemTemplateById500JSONResponse{Message: err.Error()}, nil
	}
	return DeleteChecklistItemTemplateById204Response{}, nil
}

func NewChecklistItemTemplateController(service service.IChecklistItemTemplateService) IChecklistItemTemplateController {
	return &checklistItemTemplateController{
		service: service,
		mapper:  NewChecklistItemTemplateMapper(),
	}
}
