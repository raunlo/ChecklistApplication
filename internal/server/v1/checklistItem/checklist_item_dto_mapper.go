package checklistItem

import (
	"com.raunlo.checklist/internal/core/domain"
	"github.com/rendis/structsconv"
)

type IChecklistItemDtoMapper interface {
	MapDomainToDto(checklistItem domain.ChecklistItem) ChecklistItemResponse
	MapCreateRequestToDomain(createRequest CreateChecklistItemRequest) domain.ChecklistItem
	MapUpdateRequestToDomain(updateRequest UpdateChecklistItemRequest) domain.ChecklistItem
	MapDomainListToDtoList(checklistItems []domain.ChecklistItem) []ChecklistItemResponse
	MapCreateChecklistItemRowRequestToDomain(request CreateChecklistItemRowRequest) domain.ChecklistItemRow
	MapChecklistItemRowDomainToDto(row domain.ChecklistItemRow) ChecklistItemRowResponse
}

type checklistItemMapper struct{}

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

func (mapper *checklistItemMapper) MapCreateRequestToDomain(createRequest CreateChecklistItemRequest) domain.ChecklistItem {
	checklistItem := domain.ChecklistItem{}
	structsconv.Map(&createRequest, &checklistItem)
	return checklistItem
}

func (mapper *checklistItemMapper) MapUpdateRequestToDomain(updateRequest UpdateChecklistItemRequest) domain.ChecklistItem {
	checklistItem := domain.ChecklistItem{}
	structsconv.Map(&updateRequest, &checklistItem)
	return checklistItem
}

func (mapper *checklistItemMapper) MapCreateChecklistItemRowRequestToDomain(request CreateChecklistItemRowRequest) domain.ChecklistItemRow {
	row := domain.ChecklistItemRow{}
	structsconv.Map(&request, &row)
	return row
}

func (mapper *checklistItemMapper) MapChecklistItemRowDomainToDto(row domain.ChecklistItemRow) ChecklistItemRowResponse {
	dto := ChecklistItemRowResponse{}
	structsconv.Map(&row, &dto)
	return dto
}

func NewChecklistItemMapper() IChecklistItemDtoMapper {
	return &checklistItemMapper{}
}
