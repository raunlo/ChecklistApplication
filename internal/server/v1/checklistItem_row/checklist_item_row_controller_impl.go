package checklistItem_row

import "context"

type IChecklistItemRowController = StrictServerInterface

type checklistItemRowController struct {
}

func (c *checklistItemRowController) PostApiV1ChecklistsChecklistIdItemsItemIdRows(ctx context.Context,
	request PostApiV1ChecklistsChecklistIdItemsItemIdRowsRequestObject) (PostApiV1ChecklistsChecklistIdItemsItemIdRowsResponseObject, error) {

}

func NewChecklistItemRowController() IChecklistItemRowController {
	return &checklistItemRowController{}
}
