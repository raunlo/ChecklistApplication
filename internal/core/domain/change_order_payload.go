package domain

type ChangeOrderRequest struct {
	NewOrderNumber  uint
	ChecklistId     uint
	ChecklistItemId uint
	SortOrder       SortOrder
}

type ChangeOrderResponse struct {
	OrderNumber     uint
	ChecklistItemId uint
	ChecklistId     uint
}
