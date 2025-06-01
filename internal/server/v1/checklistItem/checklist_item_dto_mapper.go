package checklistItem

import (
	"com.raunlo.checklist/internal/core/domain"
	"github.com/rendis/structsconv"
)

type IChecklistItemDtoMapper interface {
	MapDomainToDto(checklistItem domain.ChecklistItem) ChecklistItemResponse
	MapDtoToDomain(checklistItemDto CreateChecklistItemRequest) domain.ChecklistItem
	MapDomainListToDtoList(checklistItems []domain.ChecklistItem) []ChecklistItemResponse
}

type checklistItemMapper struct {
}

func (mapper *checklistItemMapper) MapDomainListToDtoList(checklistItems []domain.ChecklistItem) []ChecklistItemResponse {
	checklistItemsDtoList := make([]ChecklistItemResponse, len(checklistItems))
	for index, item := range checklistItems {
		checklistItemsDtoList[index] = mapper.MapDomainToDto(item)
	}

	return checklistItemsDtoList
}

func (mapper *checklistItemMapper) MapDomainToDto(checklistItem domain.ChecklistItem) ChecklistItemResponse {
	checklistItemDto := ChecklistItemResponse{}
	structsconv.Map(&checklistItem, &checklistItemDto)
	return checklistItemDto
}

func (mapper *checklistItemMapper) MapDtoToDomain(checklistItemDto CreateChecklistItemRequest) domain.ChecklistItem {
	checklistItem := domain.ChecklistItem{}
	structsconv.Map(&checklistItemDto, &checklistItem)
	return checklistItem
}

func NewChecklistItemMapper() IChecklistItemDtoMapper {
	return &checklistItemMapper{}
}
