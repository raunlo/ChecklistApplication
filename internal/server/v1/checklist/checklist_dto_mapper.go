package checklist

import (
	"com.raunlo.checklist/internal/core/domain"
	"github.com/rendis/structsconv"
)

type IChecklistDtoMapper interface {
	ToDomain(source CreateChecklistRequest) domain.Checklist
	ToDTO(source domain.Checklist) ChecklistResponse
	ToDtoArray(checklists []domain.Checklist) []ChecklistResponse
}

type checklistDtoMapper struct{}

func NewChecklistDtoMapper() IChecklistDtoMapper {
	return &checklistDtoMapper{}
}

func (*checklistDtoMapper) ToDomain(source CreateChecklistRequest) domain.Checklist {
	target := domain.Checklist{}
	structsconv.Map(&source, &target)
	return target
}

func (*checklistDtoMapper) ToDTO(source domain.Checklist) ChecklistResponse {
	target := ChecklistResponse{}
	structsconv.Map(&source, &target)
	return target
}

func (mapper *checklistDtoMapper) ToDtoArray(checklists []domain.Checklist) []ChecklistResponse {
	checklistDtoArray := make([]ChecklistResponse, 0)

	for _, checklist := range checklists {
		checklistDtoArray = append(checklistDtoArray, mapper.ToDTO(checklist))
	}
	return checklistDtoArray
}
