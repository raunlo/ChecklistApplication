package checklistItemTemplate

import (
	"com.raunlo.checklist/internal/core/domain"
	"github.com/rendis/structsconv"
)

type IChecklistItemTemplateDtoMapper interface {
	MapDomainToDto(template domain.ChecklistItemTemplate) ChecklistItemTemplateResponse
	MapCreateRequestToDomain(request CreateChecklistItemTemplateRequest) domain.ChecklistItemTemplate
	MapUpdateRequestToDomain(request UpdateChecklistItemTemplateRequest) domain.ChecklistItemTemplate
	MapDomainListToDtoList(templates []domain.ChecklistItemTemplate) []ChecklistItemTemplateResponse
}

type checklistItemTemplateMapper struct{}

func (m *checklistItemTemplateMapper) MapDomainToDto(template domain.ChecklistItemTemplate) ChecklistItemTemplateResponse {
	dto := ChecklistItemTemplateResponse{}
	structsconv.Map(&template, &dto)
	return dto
}

func (m *checklistItemTemplateMapper) MapCreateRequestToDomain(request CreateChecklistItemTemplateRequest) domain.ChecklistItemTemplate {
	template := domain.ChecklistItemTemplate{}
	structsconv.Map(&request, &template)
	return template
}

func (m *checklistItemTemplateMapper) MapUpdateRequestToDomain(request UpdateChecklistItemTemplateRequest) domain.ChecklistItemTemplate {
	template := domain.ChecklistItemTemplate{}
	structsconv.Map(&request, &template)
	return template
}

func (m *checklistItemTemplateMapper) MapDomainListToDtoList(templates []domain.ChecklistItemTemplate) []ChecklistItemTemplateResponse {
	res := make([]ChecklistItemTemplateResponse, len(templates))
	for i, t := range templates {
		res[i] = m.MapDomainToDto(t)
	}
	return res
}

func NewChecklistItemTemplateMapper() IChecklistItemTemplateDtoMapper {
	return &checklistItemTemplateMapper{}
}
