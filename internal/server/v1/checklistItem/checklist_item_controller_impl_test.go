package checklistItem

import (
	"context"
	"testing"

	"com.raunlo.checklist/internal/core/domain"
	service "com.raunlo.checklist/internal/core/service"
	"github.com/stretchr/testify/mock"
)

// Ensure mockChecklistItemsService implements the interface.
var _ service.IChecklistItemsService = (*mockChecklistItemsService)(nil)

type mockChecklistItemsService struct {
	mock.Mock
}

func (m *mockChecklistItemsService) SaveChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	return domain.ChecklistItem{}, nil
}

func (m *mockChecklistItemsService) UpdateChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	return domain.ChecklistItem{}, nil
}

func (m *mockChecklistItemsService) SaveChecklistItemRow(ctx context.Context, checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error) {
	args := m.Called(checklistId, itemId, row)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChecklistItemRow), err
}

func (m *mockChecklistItemsService) DeleteChecklistItemRow(ctx context.Context, checklistId uint, itemId uint, rowId uint) domain.Error {
	args := m.Called(checklistId, itemId, rowId)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

func (m *mockChecklistItemsService) FindChecklistItemById(ctx context.Context, checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
	return nil, nil
}

func (m *mockChecklistItemsService) DeleteChecklistItemById(ctx context.Context, checklistId uint, id uint) domain.Error {
	return nil
}

func (m *mockChecklistItemsService) FindAllChecklistItems(ctx context.Context, checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error) {
	return nil, nil
}

func (m *mockChecklistItemsService) ChangeChecklistItemOrder(ctx context.Context, request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error) {
	return domain.ChangeOrderResponse{}, nil
}

func (m *mockChecklistItemsService) ToggleCompleted(ctx context.Context, checklistId uint, itemId uint, completed bool) (domain.ChecklistItem, domain.Error) {
	args := m.Called(checklistId, itemId, completed)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChecklistItem), err
}

func TestChecklistItemController_CreateChecklistItemRow(t *testing.T) {
	expected := domain.ChecklistItemRow{Id: 5, Name: "row", Completed: true}
	svc := new(mockChecklistItemsService)
	svc.On("SaveChecklistItemRow", uint(1), uint(2), domain.ChecklistItemRow{Name: "row"}).Return(expected, nil)

	controller := &checklistItemController{service: svc, mapper: NewChecklistItemMapper()}
	req := CreateChecklistItemRowRequestObject{ChecklistId: 1, ItemId: 2, Body: &CreateChecklistItemRowJSONRequestBody{Name: "row"}}
	res, err := controller.CreateChecklistItemRow(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dto, ok := res.(CreateChecklistItemRow201JSONResponse)
	if !ok {
		t.Fatalf("expected CreateChecklistItemRow201JSONResponse got %T", res)
	}
	if dto.Name != expected.Name || dto.Id != expected.Id {
		t.Fatalf("unexpected dto: got %#v, want %#v", dto, expected)
	}
	if dto.Completed == nil || *dto.Completed != expected.Completed {
		t.Fatalf("unexpected completed value: got %v, want %v", dto.Completed, expected.Completed)
	}
	svc.AssertExpectations(t)
}

func TestChecklistItemController_CreateChecklistItemRow_NotFound(t *testing.T) {
	svc := new(mockChecklistItemsService)
	svc.On("SaveChecklistItemRow", uint(1), uint(2), domain.ChecklistItemRow{Name: "row"}).Return(domain.ChecklistItemRow{}, domain.NewError("missing", 404))

	controller := &checklistItemController{service: svc, mapper: NewChecklistItemMapper()}
	req := CreateChecklistItemRowRequestObject{ChecklistId: 1, ItemId: 2, Body: &CreateChecklistItemRowJSONRequestBody{Name: "row"}}
	res, err := controller.CreateChecklistItemRow(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.(CreateChecklistItemRow404JSONResponse); !ok {
		t.Fatalf("expected CreateChecklistItemRow404JSONResponse got %T", res)
	}
	svc.AssertExpectations(t)
}

func TestChecklistItemController_DeleteChecklistItemRow(t *testing.T) {
	svc := new(mockChecklistItemsService)
	svc.On("DeleteChecklistItemRow", uint(1), uint(2), uint(3)).Return(nil)

	controller := &checklistItemController{service: svc, mapper: NewChecklistItemMapper()}
	req := DeleteChecklistItemRowRequestObject{ChecklistId: 1, ItemId: 2, RowId: 3}
	res, err := controller.DeleteChecklistItemRow(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.(DeleteChecklistItemRow204Response); !ok {
		t.Fatalf("expected DeleteChecklistItemRow204Response got %T", res)
	}
	svc.AssertExpectations(t)
}

func TestChecklistItemController_DeleteChecklistItemRow_NotFound(t *testing.T) {
	svc := new(mockChecklistItemsService)
	svc.On("DeleteChecklistItemRow", uint(1), uint(2), uint(3)).Return(domain.NewError("missing", 404))

	controller := &checklistItemController{service: svc, mapper: NewChecklistItemMapper()}
	req := DeleteChecklistItemRowRequestObject{ChecklistId: 1, ItemId: 2, RowId: 3}
	res, err := controller.DeleteChecklistItemRow(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.(DeleteChecklistItemRow404JSONResponse); !ok {
		t.Fatalf("expected DeleteChecklistItemRow404JSONResponse got %T", res)
	}
	svc.AssertExpectations(t)
}
