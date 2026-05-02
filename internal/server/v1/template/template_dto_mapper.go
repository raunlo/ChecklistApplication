package template

import (
	"com.raunlo.checklist/internal/core/domain"
	"github.com/rendis/structsconv"
)

type ITemplateDtoMapper interface {
	ToDomain(source CreateTemplateRequest) domain.Template
	ToDTO(source domain.Template) TemplateResponse
	ToTemplateDtoArray(templates []domain.Template) []TemplateResponse
	ToTemplateRowDTO(source domain.TemplateRow) TemplateRowResponse
	ToTemplateRowDtoArray(rows []domain.TemplateRow) []TemplateRowResponse
}

type templateDtoMapper struct{}

func NewTemplateDtoMapper() ITemplateDtoMapper {
	return &templateDtoMapper{}
}

func (*templateDtoMapper) ToDomain(source CreateTemplateRequest) domain.Template {
	target := domain.Template{}
	structsconv.Map(&source, &target)

	if source.Rows != nil {
		rows := make([]domain.TemplateRow, len(*source.Rows))
		for i, r := range *source.Rows {
			rows[i] = domain.TemplateRow{
				Name:     r.Name,
				Position: r.Position,
			}
		}
		target.Rows = rows
	}

	return target
}

func (mapper *templateDtoMapper) ToDTO(source domain.Template) TemplateResponse {
	target := TemplateResponse{}
	structsconv.Map(&source, &target)

	target.Rows = mapper.ToTemplateRowDtoArray(source.Rows)
	target.IsOwner = source.IsOwner
	target.WorkspaceIds = make([]uint, len(source.WorkspaceIds))
	copy(target.WorkspaceIds, source.WorkspaceIds)

	return target
}

func (mapper *templateDtoMapper) ToTemplateDtoArray(templates []domain.Template) []TemplateResponse {
	templateDtoArray := make([]TemplateResponse, 0)

	for _, t := range templates {
		templateDtoArray = append(templateDtoArray, mapper.ToDTO(t))
	}

	return templateDtoArray
}

func (*templateDtoMapper) ToTemplateRowDTO(source domain.TemplateRow) TemplateRowResponse {
	target := TemplateRowResponse{}
	structsconv.Map(&source, &target)
	return target
}

func (mapper *templateDtoMapper) ToTemplateRowDtoArray(rows []domain.TemplateRow) []TemplateRowResponse {
	rowDtoArray := make([]TemplateRowResponse, 0)

	for _, row := range rows {
		rowDtoArray = append(rowDtoArray, mapper.ToTemplateRowDTO(row))
	}

	return rowDtoArray
}
