package service

import (
	"context"
	"testing"
	"time"

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

// mockChecklistOwnershipChecker uses testify's mock for guardrail.IChecklistOwnershipChecker.
type mockChecklistOwnershipChecker struct {
	mock.Mock
}

func (m *mockChecklistOwnershipChecker) HasAccessToChecklist(ctx context.Context, checklistId uint) domain.Error {
	args := m.Called(ctx, checklistId)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

func (m *mockChecklistOwnershipChecker) IsChecklistOwner(ctx context.Context, checklistId uint) domain.Error {
	args := m.Called(ctx, checklistId)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
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

func (m *mockNotificationService) NotifyItemSoftDeleted(ctx context.Context, checklistId uint, itemId uint) {
	m.Called(ctx, checklistId, itemId)
}

func (m *mockNotificationService) NotifyItemRestored(ctx context.Context, checklistId uint, item domain.ChecklistItem) {
	m.Called(ctx, checklistId, item)
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

func (m *mockChecklistItemsRepository) DeleteChecklistItemRowAndAutoComplete(ctx context.Context, checklistId uint, itemId uint, rowId uint) (domain.ChecklistItemRowDeletionResult, domain.Error) {
	args := m.Called(ctx, checklistId, itemId, rowId)
	var result domain.ChecklistItemRowDeletionResult
	var err domain.Error
	if arg := args.Get(0); arg != nil {
		result = arg.(domain.ChecklistItemRowDeletionResult)
	}
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return result, err
}

func (m *mockChecklistItemsRepository) FindChecklistItemById(ctx context.Context, checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
	args := m.Called(ctx, checklistId, id)
	var item *domain.ChecklistItem
	var err domain.Error
	if arg := args.Get(0); arg != nil {
		item = arg.(*domain.ChecklistItem)
	}
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return item, err
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

func (m *mockChecklistItemsRepository) RebalancePositions(ctx context.Context, checklistId uint) domain.Error {
	args := m.Called(ctx, checklistId)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

func (m *mockChecklistItemsRepository) RestoreChecklistItem(ctx context.Context, checklistId uint, itemId uint) (domain.ChecklistItem, domain.Error) {
	args := m.Called(ctx, checklistId, itemId)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChecklistItem), err
}

func (m *mockChecklistItemsRepository) PurgeSoftDeletedItems(ctx context.Context, retentionPeriod time.Duration) (int64, domain.Error) {
	args := m.Called(ctx, retentionPeriod)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(int64), err
}

func (m *mockChecklistItemsRepository) TryAcquireCleanupLock(ctx context.Context, minInterval time.Duration) (bool, domain.Error) {
	args := m.Called(ctx, minInterval)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(bool), err
}

func (m *mockChecklistItemsRepository) ReleaseCleanupLock(ctx context.Context) domain.Error {
	args := m.Called(ctx)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

func (m *mockChecklistItemsRepository) UpdateCleanupLastRun(ctx context.Context) domain.Error {
	args := m.Called(ctx)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

func TestChecklistItemsService_SaveChecklistItemRow(t *testing.T) {
	expected := domain.ChecklistItemRow{Id: 1, Name: "row", Completed: false}
	repo := new(mockChecklistItemsRepository)
	notifier := new(mockNotificationService)
	ownershipChecker := new(mockChecklistOwnershipChecker)
	repo.On("SaveChecklistItemRow", mock.Anything, uint(10), uint(20), domain.ChecklistItemRow{Name: "row"}).Return(expected, nil)
	notifier.On("NotifyItemRowAdded", mock.Anything, uint(10), uint(20), expected).Return()
	ownershipChecker.On("HasAccessToChecklist", mock.Anything, uint(10)).Return(nil)

	svc := &checklistItemsService{repository: repo, notifier: notifier, checklistOwnershipChecker: ownershipChecker}
	row, err := svc.SaveChecklistItemRow(context.Background(), 10, 20, domain.ChecklistItemRow{Name: "row"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row != expected {
		t.Fatalf("expected %#v got %#v", expected, row)
	}
	repo.AssertExpectations(t)
	notifier.AssertExpectations(t)
	ownershipChecker.AssertExpectations(t)
}

func TestChecklistItemsService_SaveChecklistItemRow_Error(t *testing.T) {
	expectedErr := domain.NewError("fail", 500)
	repo := new(mockChecklistItemsRepository)
	notifier := new(mockNotificationService)
	ownershipChecker := new(mockChecklistOwnershipChecker)
	repo.On("SaveChecklistItemRow", mock.Anything, uint(1), uint(2), domain.ChecklistItemRow{Name: "x"}).Return(domain.ChecklistItemRow{}, expectedErr)
	ownershipChecker.On("HasAccessToChecklist", mock.Anything, uint(1)).Return(nil)

	svc := &checklistItemsService{repository: repo, notifier: notifier, checklistOwnershipChecker: ownershipChecker}
	_, err := svc.SaveChecklistItemRow(context.Background(), 1, 2, domain.ChecklistItemRow{Name: "x"})
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
	ownershipChecker := new(mockChecklistOwnershipChecker)
	repo.On("DeleteChecklistItemRowAndAutoComplete", mock.Anything, uint(1), uint(2), uint(3)).Return(
		domain.ChecklistItemRowDeletionResult{Success: true, ItemAutoCompleted: false},
		nil,
	)
	notifier.On("NotifyItemRowDeleted", mock.Anything, uint(1), uint(2), uint(3)).Return()
	ownershipChecker.On("HasAccessToChecklist", mock.Anything, uint(1)).Return(nil)

	svc := &checklistItemsService{repository: repo, notifier: notifier, checklistOwnershipChecker: ownershipChecker}
	err := svc.DeleteChecklistItemRow(context.Background(), 1, 2, 3)
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
	ownershipChecker := new(mockChecklistOwnershipChecker)
	repo.On("DeleteChecklistItemRowAndAutoComplete", mock.Anything, uint(1), uint(2), uint(3)).Return(
		domain.ChecklistItemRowDeletionResult{Success: false, ItemAutoCompleted: false},
		expectedErr,
	)
	ownershipChecker.On("HasAccessToChecklist", mock.Anything, uint(1)).Return(nil)

	svc := &checklistItemsService{repository: repo, notifier: notifier, checklistOwnershipChecker: ownershipChecker}
	err := svc.DeleteChecklistItemRow(context.Background(), 1, 2, 3)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != expectedErr {
		t.Fatalf("expected %v got %v", expectedErr, err)
	}
	repo.AssertExpectations(t)
}

func TestChecklistItemsService_RestoreChecklistItem_Success(t *testing.T) {
	expectedItem := domain.ChecklistItem{
		Id:        1,
		Name:      "Restored Item",
		Completed: false,
		Position:  1.0,
		Rows: []domain.ChecklistItemRow{
			{Id: 10, Name: "Subitem", Completed: false},
		},
	}
	repo := new(mockChecklistItemsRepository)
	notifier := new(mockNotificationService)
	ownershipChecker := new(mockChecklistOwnershipChecker)

	ownershipChecker.On("HasAccessToChecklist", mock.Anything, uint(100)).Return(nil)
	repo.On("RestoreChecklistItem", mock.Anything, uint(100), uint(1)).Return(expectedItem, nil)
	notifier.On("NotifyItemRestored", mock.Anything, uint(100), expectedItem).Return()

	svc := &checklistItemsService{repository: repo, notifier: notifier, checklistOwnershipChecker: ownershipChecker}
	item, err := svc.RestoreChecklistItem(context.Background(), 100, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Id != expectedItem.Id {
		t.Fatalf("expected item id %d got %d", expectedItem.Id, item.Id)
	}
	if item.Name != expectedItem.Name {
		t.Fatalf("expected item name %s got %s", expectedItem.Name, item.Name)
	}
	if len(item.Rows) != 1 {
		t.Fatalf("expected 1 row got %d", len(item.Rows))
	}
	repo.AssertExpectations(t)
	notifier.AssertExpectations(t)
	ownershipChecker.AssertExpectations(t)
}

func TestChecklistItemsService_RestoreChecklistItem_NotFound(t *testing.T) {
	expectedErr := domain.NewError("item not found or not deleted", 404)
	repo := new(mockChecklistItemsRepository)
	notifier := new(mockNotificationService)
	ownershipChecker := new(mockChecklistOwnershipChecker)

	ownershipChecker.On("HasAccessToChecklist", mock.Anything, uint(100)).Return(nil)
	repo.On("RestoreChecklistItem", mock.Anything, uint(100), uint(999)).Return(domain.ChecklistItem{}, expectedErr)

	svc := &checklistItemsService{repository: repo, notifier: notifier, checklistOwnershipChecker: ownershipChecker}
	_, err := svc.RestoreChecklistItem(context.Background(), 100, 999)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.ResponseCode() != 404 {
		t.Fatalf("expected 404 got %d", err.ResponseCode())
	}
	repo.AssertExpectations(t)
	// Notifier should NOT be called on error
	notifier.AssertNotCalled(t, "NotifyItemRestored", mock.Anything, mock.Anything, mock.Anything)
}

func TestChecklistItemsService_RestoreChecklistItem_AccessDenied(t *testing.T) {
	accessErr := domain.NewError("access denied", 403)
	repo := new(mockChecklistItemsRepository)
	notifier := new(mockNotificationService)
	ownershipChecker := new(mockChecklistOwnershipChecker)

	ownershipChecker.On("HasAccessToChecklist", mock.Anything, uint(100)).Return(accessErr)

	svc := &checklistItemsService{repository: repo, notifier: notifier, checklistOwnershipChecker: ownershipChecker}
	_, err := svc.RestoreChecklistItem(context.Background(), 100, 1)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.ResponseCode() != 403 {
		t.Fatalf("expected 403 got %d", err.ResponseCode())
	}
	// Repository should NOT be called if access denied
	repo.AssertNotCalled(t, "RestoreChecklistItem", mock.Anything, mock.Anything, mock.Anything)
	notifier.AssertNotCalled(t, "NotifyItemRestored", mock.Anything, mock.Anything, mock.Anything)
}
