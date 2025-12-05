package service

import (
	"context"
	"testing"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

// mockChecklistItemsRepository uses testify's mock for repository.IChecklistItemsRepository.
type mockChecklistItemsRepository struct {
	mock.Mock
}

// mockNotificationService uses testify's mock for notification.INotificationService.
type mockNotificationService struct {
	mock.Mock
}

func (m *mockNotificationService) NotifyItemCreated(ctx context.Context, checklistId uint, item domain.ChecklistItem) {
	m.Called(ctx, checklistId, item)
}

func (m *mockNotificationService) NotifyItemUpdated(ctx context.Context, checklistId uint, item domain.ChecklistItem) {
	m.Called(ctx, checklistId, item)
}

func (m *mockNotificationService) NotifyItemDeleted(ctx context.Context, checklistId uint, itemId uint) {
	m.Called(ctx, checklistId, itemId)
}

func (m *mockNotificationService) NotifyItemRowAdded(ctx context.Context, checklistId uint, itemId uint, row domain.ChecklistItemRow) {
	m.Called(ctx, checklistId, itemId, row)
}

func (m *mockNotificationService) NotifyItemRowDeleted(ctx context.Context, checklistId uint, itemId uint, rowId uint) {
	m.Called(ctx, checklistId, itemId, rowId)
}

func (m *mockNotificationService) NotifyItemReordered(ctx context.Context, request domain.ChangeOrderRequest, resp domain.ChangeOrderResponse) {
	m.Called(ctx, request, resp)
}

func (m *mockChecklistItemsRepository) UpdateChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	return domain.ChecklistItem{}, nil
}

func (m *mockChecklistItemsRepository) SaveChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	return domain.ChecklistItem{}, nil
}

func (m *mockChecklistItemsRepository) SaveChecklistItemRow(ctx context.Context, checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error) {
	args := m.Called(ctx, checklistId, itemId, row)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChecklistItemRow), err
}

func (m *mockChecklistItemsRepository) ToggleItemCompleted(ctx context.Context, checklistId uint, checklistItemId uint, completed bool) (domain.ChecklistItem, domain.Error) {
	args := m.Called(ctx, checklistId, checklistItemId, completed)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChecklistItem), err
}

func (m *mockChecklistItemsRepository) DeleteChecklistItemRow(ctx context.Context, checklistId uint, itemId uint, rowId uint) domain.Error {
	args := m.Called(ctx, checklistId, itemId, rowId)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

func (m *mockChecklistItemsRepository) DeleteChecklistItemRowAndAutoComplete(ctx context.Context, checklistId uint, itemId uint, rowId uint) domain.Error {
	args := m.Called(ctx, checklistId, itemId, rowId)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

func (m *mockChecklistItemsRepository) FindChecklistItemById(ctx context.Context, checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
	return nil, nil
}

func (m *mockChecklistItemsRepository) DeleteChecklistItemById(ctx context.Context, checklistId uint, id uint) domain.Error {
	return nil
}

func (m *mockChecklistItemsRepository) FindAllChecklistItems(ctx context.Context, checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error) {
	return nil, nil
}

func (m *mockChecklistItemsRepository) ChangeChecklistItemOrder(ctx context.Context, request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error) {
	return domain.ChangeOrderResponse{}, nil
}

func TestChecklistItemsService_SaveChecklistItemRow(t *testing.T) {
	expected := domain.ChecklistItemRow{Id: 1, Name: "row", Completed: false}
	repo := new(mockChecklistItemsRepository)
	notifier := new(mockNotificationService)
	repo.On("SaveChecklistItemRow", mock.Anything, uint(10), uint(20), domain.ChecklistItemRow{Name: "row"}).Return(expected, nil)
	notifier.On("NotifyItemRowAdded", mock.Anything, uint(10), uint(20), expected).Return()

	svc := &checklistItemsService{repository: repo, notifier: notifier}
	row, err := svc.SaveChecklistItemRow(t.Context(), 10, 20, domain.ChecklistItemRow{Name: "row"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row != expected {
		t.Fatalf("expected %#v got %#v", expected, row)
	}
	repo.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

func TestChecklistItemsService_SaveChecklistItemRow_Error(t *testing.T) {
	expectedErr := domain.NewError("fail", 500)
	repo := new(mockChecklistItemsRepository)
	notifier := new(mockNotificationService)
	repo.On("SaveChecklistItemRow", mock.Anything, uint(1), uint(2), domain.ChecklistItemRow{Name: "x"}).Return(domain.ChecklistItemRow{}, expectedErr)

	svc := &checklistItemsService{repository: repo, notifier: notifier}
	_, err := svc.SaveChecklistItemRow(t.Context(), 1, 2, domain.ChecklistItemRow{Name: "x"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != expectedErr {
		t.Fatalf("expected error %v got %v", expectedErr, err)
	}
	repo.AssertExpectations(t)
}

func TestChecklistItemsService_DeleteChecklistItemRow(t *testing.T) {
	repo := new(mockChecklistItemsRepository)
	notifier := new(mockNotificationService)
	repo.On("DeleteChecklistItemRowAndAutoComplete", mock.Anything, uint(1), uint(2), uint(3)).Return(nil)
	repo.On("FindChecklistItemById", mock.Anything, uint(1), uint(2)).Return(&domain.ChecklistItem{Id: 2, Completed: false}, nil)
	notifier.On("NotifyItemRowDeleted", mock.Anything, uint(1), uint(2), uint(3)).Return()

	svc := &checklistItemsService{repository: repo, notifier: notifier}
	err := svc.DeleteChecklistItemRow(t.Context(), 1, 2, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	repo.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

func TestChecklistItemsService_DeleteChecklistItemRow_Error(t *testing.T) {
	expectedErr := domain.NewError("missing", 404)
	repo := new(mockChecklistItemsRepository)
	notifier := new(mockNotificationService)
	repo.On("DeleteChecklistItemRowAndAutoComplete", mock.Anything, uint(1), uint(2), uint(3)).Return(expectedErr)

	svc := &checklistItemsService{repository: repo, notifier: notifier}
	err := svc.DeleteChecklistItemRow(t.Context(), 1, 2, 3)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != expectedErr {
		t.Fatalf("expected %v got %v", expectedErr, err)
	}
	repo.AssertExpectations(t)
}
