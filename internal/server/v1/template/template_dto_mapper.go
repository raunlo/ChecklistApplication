package template

import (
	"com.raunlo.checklist/internal/core/domain"
	"github.com/rendis/structsconv"
)

type ITemplateDtoMapper interface {
	ToDomain(source CreateTemplateRequest) domain.Template
	ToDTO(source domain.Template) TemplateResponse
	ToTemplateDtoArray(templates []domain.Template) []TemplateResponse
	ToTemplateItemDTO(source domain.TemplateItem) TemplateItemResponse
	ToTemplateItemDtoArray(items []domain.TemplateItem) []TemplateItemResponse
	ToTemplatePreviewDTO(existingItems []domain.TemplateItem, newItems []domain.TemplateItem) TemplatePreviewResponse
}

type templateDtoMapper struct{}

func NewTemplateDtoMapper() ITemplateDtoMapper {
	return &templateDtoMapper{}
}

func (*templateDtoMapper) ToDomain(source CreateTemplateRequest) domain.Template {
	target := domain.Template{}
	structsconv.Map(&source, &target)
	return target
}

func (mapper *templateDtoMapper) ToDTO(source domain.Template) TemplateResponse {
	target := TemplateResponse{}
	structsconv.Map(&source, &target)

	// Map items array
	target.Items = mapper.ToTemplateItemDtoArray(source.Items)

	return target
}

func (mapper *templateDtoMapper) ToTemplateDtoArray(templates []domain.Template) []TemplateResponse {
	templateDtoArray := make([]TemplateResponse, 0)

	for _, template := range templates {
		templateDtoArray = append(templateDtoArray, mapper.ToDTO(template))
	}

	return templateDtoArray
}

func (*templateDtoMapper) ToTemplateItemDTO(source domain.TemplateItem) TemplateItemResponse {
	target := TemplateItemResponse{}
	structsconv.Map(&source, &target)
	return target
}

func (mapper *templateDtoMapper) ToTemplateItemDtoArray(items []domain.TemplateItem) []TemplateItemResponse {
	itemDtoArray := make([]TemplateItemResponse, 0)

	for _, item := range items {
		itemDtoArray = append(itemDtoArray, mapper.ToTemplateItemDTO(item))
	}

	return itemDtoArray
}

func (mapper *templateDtoMapper) ToTemplatePreviewDTO(existingItems []domain.TemplateItem, newItems []domain.TemplateItem) TemplatePreviewResponse {
	return TemplatePreviewResponse{
		ExistingItems: mapper.ToTemplateItemDtoArray(existingItems),
		NewItems:      mapper.ToTemplateItemDtoArray(newItems),
	}
}
