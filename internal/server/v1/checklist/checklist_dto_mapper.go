package checklist

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/rendis/structsconv"
)

type IChecklistDtoMapper interface {
	ToDomain(source CreateChecklistRequest) domain.Checklist
	ToDTO(source domain.Checklist, ctx context.Context) ChecklistResponse
	ToDtoArray(checklists []domain.Checklist, ctx context.Context) []ChecklistResponse
	ToChecklistListResponseWithStats(source []domain.Checklist, ctx context.Context) GetChecklistsWithStatsResponse
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

func (*checklistDtoMapper) ToDTO(source domain.Checklist, ctx context.Context) ChecklistResponse {
	target := ChecklistResponse{}
	structsconv.Map(&source, &target)

	// Get current user ID from context
	currentUserId, _ := domain.GetUserIdFromContext(ctx)

	// Use stats from domain object (calculated in SQL)
	target.Stats.TotalItems = source.Stats.TotalItems
	target.Stats.CompletedItems = source.Stats.CompletedItems

	// Set owner information
	target.Owner = source.Owner
	target.IsOwner = (source.Owner == currentUserId)
	target.IsShared = (len(source.SharedWith) > 0)

	// Only show SharedWith array if current user is the owner
	if target.IsOwner {
		target.SharedWith = &source.SharedWith
	}

	// Don't include items array in list view (only in detail view)
	// Items will be nil by default

	return target
}

func (mapper *checklistDtoMapper) ToDtoArray(checklists []domain.Checklist, ctx context.Context) []ChecklistResponse {
	checklistDtoArray := make([]ChecklistResponse, 0)

	for _, checklist := range checklists {
		checklistDtoArray = append(checklistDtoArray, mapper.ToDTO(checklist, ctx))
	}
	return checklistDtoArray
}

func (mapper *checklistDtoMapper) ToChecklistListResponseWithStats(source []domain.Checklist, ctx context.Context) GetChecklistsWithStatsResponse {
	response := make([]ChecklistWithStats, 0)

	for _, checklist := range source {
		dto := ChecklistWithStats{}

		dto.Id = checklist.Id
		dto.Name = checklist.Name
		dto.Stats.CompletedItems = checklist.Stats.CompletedItems
		dto.Stats.TotalItems = checklist.Stats.TotalItems
		// Get current user ID from context
		currentUserId, _ := domain.GetUserIdFromContext(ctx)

		dto.IsOwner = (checklist.Owner == currentUserId)
		dto.IsShared = (len(checklist.SharedWith) > 0)

		response = append(response, dto)
	}

	return GetChecklistsWithStatsResponse{Checklists: response}
}
