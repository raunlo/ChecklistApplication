package service

import (
	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/repository"
	"com.raunlo.checklist/internal/sse"
)

type IChecklistItemsService interface {
	SaveChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	UpdateChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error)
	SaveChecklistItemRow(checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error)
	FindChecklistItemById(checklistId uint, id uint) (*domain.ChecklistItem, domain.Error)
	DeleteChecklistItemById(checklistId uint, id uint) domain.Error
	DeleteChecklistItemRow(checklistId uint, itemId uint, rowId uint) domain.Error
	FindAllChecklistItems(checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error)
	ChangeChecklistItemOrder(request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error)
	ToggleCompleted(checklistId uint, itemId uint, completed bool) (domain.ChecklistItem, domain.Error)
}

type checklistItemsService struct {
	repository      repository.IChecklistItemsRepository
	templateService IChecklistItemTemplateService
}

func (service *checklistItemsService) UpdateChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	//checklistItem.Completed = isChecklistItemCompleted(checklistItem)
	result, err := service.repository.UpdateChecklistItem(checklistId, checklistItem)
	if err == nil {
		checklistIdInt := int(checklistId)
		// Publish via SSE (typed)
		sse.PublishEvent(sse.Event{
			Type:        "itemUpdated",
			ChecklistID: checklistIdInt,
			Payload:     result,
		})
	}
	return result, err
}

func (service *checklistItemsService) SaveChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	//checklistItem.Completed = isChecklistItemCompleted(checklistItem)
	result, err := service.repository.SaveChecklistItem(checklistId, checklistItem)
	if err == nil {
		checklistIdInt := int(checklistId)
		sse.PublishEvent(sse.Event{
			Type:        "itemCreated",
			ChecklistID: checklistIdInt,
			Payload:     result,
		})
	}
	return result, err
}

func (service *checklistItemsService) SaveChecklistItemRow(checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error) {
	result, err := service.repository.SaveChecklistItemRow(checklistId, itemId, row)
	if err == nil {
		checklistIdInt := int(checklistId)
		// For row operations, we'll send an update message for the entire item via SSE
		sse.PublishEvent(sse.Event{
			Type:        "itemUpdated",
			ChecklistID: checklistIdInt,
			Payload: map[string]interface{}{
				"row":         result,
				"itemId":      itemId,
				"checklistId": checklistId,
			},
		})
	}
	return result, err
}

func (service *checklistItemsService) FindChecklistItemById(checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
	return service.repository.FindChecklistItemById(checklistId, id)
}

func (service *checklistItemsService) DeleteChecklistItemById(checklistId uint, id uint) domain.Error {
	err := service.repository.DeleteChecklistItemById(checklistId, id)
	if err == nil {
		checklistIdInt := int(checklistId)
		sse.PublishEvent(sse.Event{
			Type:        "itemDeleted",
			ChecklistID: checklistIdInt,
			Payload:     map[string]interface{}{"itemId": id},
		})
	}
	return err
}

func (service *checklistItemsService) DeleteChecklistItemRow(checklistId uint, itemId uint, rowId uint) domain.Error {
	err := service.repository.DeleteChecklistItemRow(checklistId, itemId, rowId)
	if err == nil {
		checklistIdInt := int(checklistId)
		sse.PublishEvent(sse.Event{
			Type:        "itemUpdated",
			ChecklistID: checklistIdInt,
			Payload: map[string]interface{}{
				"rowId":       rowId,
				"itemId":      itemId,
				"checklistId": checklistId,
			},
		})
	}
	return err
}

func (service *checklistItemsService) FindAllChecklistItems(checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error) {
	return service.repository.FindAllChecklistItems(checklistId, completed, sortOrder)
}

func (service *checklistItemsService) ChangeChecklistItemOrder(request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error) {
	result, err := service.repository.ChangeChecklistItemOrder(request)
	if err == nil {
		checklistIdInt := int(request.ChecklistId)
		// Broadcast order change via SSE
		sse.PublishEvent(sse.Event{
			Type:        "itemReordered",
			ChecklistID: checklistIdInt,
			Payload: map[string]interface{}{
				"itemId":         request.ChecklistItemId,
				"newOrderNumber": result.OrderNumber,
				"orderChanged":   true,
			},
		})
	}
	return result, err
}

func isChecklistItemCompleted(checklistItem domain.ChecklistItem) bool {
	for _, row := range checklistItem.Rows {
		if !row.Completed {
			return false
		}
	}
	return true
}

func (service *checklistItemsService) ToggleCompleted(checklistId uint, itemId uint, completed bool) (domain.ChecklistItem, domain.Error) {
	result, err := service.repository.ToggleItemCompleted(checklistId, itemId, completed)
	if err == nil {
		checklistIdInt := int(checklistId)
		// Broadcast completion toggle via SSE
		sse.PublishEvent(sse.Event{
			Type:        "itemToggled",
			ChecklistID: checklistIdInt,
			Payload: map[string]interface{}{
				"itemId":    itemId,
				"completed": completed,
				"toggled":   true,
			},
		})
	}
	return result, err
}
