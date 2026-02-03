package checklistItem

import (
	"context"
	"net/http/httptest"
	"testing"

	"com.raunlo.checklist/internal/core/domain"
	service "com.raunlo.checklist/internal/core/service"
	"github.com/gin-gonic/gin"
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
	args := m.Called(ctx, checklistId, itemId, row)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChecklistItemRow), err
}

func (m *mockChecklistItemsService) DeleteChecklistItemRow(ctx context.Context, checklistId uint, itemId uint, rowId uint) domain.Error {
	args := m.Called(ctx, checklistId, itemId, rowId)
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

func (m *mockChecklistItemsService) RestoreChecklistItem(ctx context.Context, checklistId uint, itemId uint) (domain.ChecklistItem, domain.Error) {
	args := m.Called(ctx, checklistId, itemId)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChecklistItem), err
}

// createTestGinContext creates a gin.Context for testing
func createTestGinContext() *gin.Context {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c
}

func TestChecklistItemController_CreateChecklistItemRow(t *testing.T) {
	expected := domain.ChecklistItemRow{Id: 5, Name: "row", Completed: true}
	svc := new(mockChecklistItemsService)
	svc.On("SaveChecklistItemRow", mock.Anything, uint(1), uint(2), domain.ChecklistItemRow{Name: "row"}).Return(expected, nil)

	controller := &checklistItemController{service: svc, mapper: NewChecklistItemMapper()}
	req := CreateChecklistItemRowRequestObject{ChecklistId: 1, ItemId: 2, Body: &CreateChecklistItemRowJSONRequestBody{Name: "row"}}
	res, err := controller.CreateChecklistItemRow(createTestGinContext(), req)
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
	svc.On("SaveChecklistItemRow", mock.Anything, uint(1), uint(2), domain.ChecklistItemRow{Name: "row"}).Return(domain.ChecklistItemRow{}, domain.NewError("missing", 404))

	controller := &checklistItemController{service: svc, mapper: NewChecklistItemMapper()}
	req := CreateChecklistItemRowRequestObject{ChecklistId: 1, ItemId: 2, Body: &CreateChecklistItemRowJSONRequestBody{Name: "row"}}
	res, err := controller.CreateChecklistItemRow(createTestGinContext(), req)
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
	svc.On("DeleteChecklistItemRow", mock.Anything, uint(1), uint(2), uint(3)).Return(nil)

	controller := &checklistItemController{service: svc, mapper: NewChecklistItemMapper()}
	req := DeleteChecklistItemRowRequestObject{ChecklistId: 1, ItemId: 2, RowId: 3}
	res, err := controller.DeleteChecklistItemRow(createTestGinContext(), req)
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
	svc.On("DeleteChecklistItemRow", mock.Anything, uint(1), uint(2), uint(3)).Return(domain.NewError("missing", 404))

	controller := &checklistItemController{service: svc, mapper: NewChecklistItemMapper()}
	req := DeleteChecklistItemRowRequestObject{ChecklistId: 1, ItemId: 2, RowId: 3}
	res, err := controller.DeleteChecklistItemRow(createTestGinContext(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.(DeleteChecklistItemRow404JSONResponse); !ok {
		t.Fatalf("expected DeleteChecklistItemRow404JSONResponse got %T", res)
	}
	svc.AssertExpectations(t)
}

func TestChecklistItemController_RestoreChecklistItem_Success(t *testing.T) {
	expectedItem := domain.ChecklistItem{
		Id:        5,
		Name:      "Restored Item",
		Completed: false,
		Position:  1.0,
		Rows: []domain.ChecklistItemRow{
			{Id: 10, Name: "Subitem", Completed: false},
		},
	}
	svc := new(mockChecklistItemsService)
	svc.On("RestoreChecklistItem", mock.Anything, uint(1), uint(5)).Return(expectedItem, nil)

	controller := &checklistItemController{service: svc, mapper: NewChecklistItemMapper()}
	req := RestoreChecklistItemRequestObject{ChecklistId: 1, ItemId: 5}
	res, err := controller.RestoreChecklistItem(createTestGinContext(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	jsonRes, ok := res.(RestoreChecklistItem200JSONResponse)
	if !ok {
		t.Fatalf("expected RestoreChecklistItem200JSONResponse got %T", res)
	}
	if jsonRes.Id != 5 {
		t.Fatalf("expected id 5 got %d", jsonRes.Id)
	}
	if jsonRes.Name != "Restored Item" {
		t.Fatalf("expected name 'Restored Item' got %s", jsonRes.Name)
	}
	if jsonRes.Rows == nil || len(jsonRes.Rows) != 1 {
		t.Fatalf("expected 1 row got %v", jsonRes.Rows)
	}
	svc.AssertExpectations(t)
}

func TestChecklistItemController_RestoreChecklistItem_NotFound(t *testing.T) {
	svc := new(mockChecklistItemsService)
	svc.On("RestoreChecklistItem", mock.Anything, uint(1), uint(999)).Return(domain.ChecklistItem{}, domain.NewError("not found", 404))

	controller := &checklistItemController{service: svc, mapper: NewChecklistItemMapper()}
	req := RestoreChecklistItemRequestObject{ChecklistId: 1, ItemId: 999}
	res, err := controller.RestoreChecklistItem(createTestGinContext(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.(RestoreChecklistItem404JSONResponse); !ok {
		t.Fatalf("expected RestoreChecklistItem404JSONResponse got %T", res)
	}
	svc.AssertExpectations(t)
}

func TestChecklistItemController_RestoreChecklistItem_ServerError(t *testing.T) {
	svc := new(mockChecklistItemsService)
	svc.On("RestoreChecklistItem", mock.Anything, uint(1), uint(5)).Return(domain.ChecklistItem{}, domain.NewError("db error", 500))

	controller := &checklistItemController{service: svc, mapper: NewChecklistItemMapper()}
	req := RestoreChecklistItemRequestObject{ChecklistId: 1, ItemId: 5}
	res, err := controller.RestoreChecklistItem(createTestGinContext(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res.(RestoreChecklistItem500JSONResponse); !ok {
		t.Fatalf("expected RestoreChecklistItem500JSONResponse got %T", res)
	}
	svc.AssertExpectations(t)
}
