package service

import (
	"context"
	"testing"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

// mockChecklistRepository uses testify's mock for repository.IChecklistRepository.
type mockChecklistRepository struct {
	mock.Mock
}

func (m *mockChecklistRepository) UpdateChecklist(ctx context.Context, checklist domain.Checklist) (domain.Checklist, domain.Error) {
	args := m.Called(ctx, checklist)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.Checklist), err
}

func (m *mockChecklistRepository) SaveChecklist(ctx context.Context, checklist domain.Checklist) (domain.Checklist, domain.Error) {
	args := m.Called(ctx, checklist)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.Checklist), err
}

func (m *mockChecklistRepository) FindChecklistById(ctx context.Context, id uint) (*domain.Checklist, domain.Error) {
	args := m.Called(ctx, id)
	var checklist *domain.Checklist
	var err domain.Error
	if arg := args.Get(0); arg != nil {
		checklist = arg.(*domain.Checklist)
	}
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return checklist, err
}

func (m *mockChecklistRepository) DeleteChecklistById(ctx context.Context, id uint) domain.Error {
	args := m.Called(ctx, id)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

func (m *mockChecklistRepository) FindAllChecklists(ctx context.Context) ([]domain.Checklist, domain.Error) {
	args := m.Called(ctx)
	var checklists []domain.Checklist
	var err domain.Error
	if arg := args.Get(0); arg != nil {
		checklists = arg.([]domain.Checklist)
	}
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return checklists, err
}

func (m *mockChecklistRepository) LeaveSharedChecklist(ctx context.Context, checklistId uint) domain.Error {
	args := m.Called(ctx, checklistId)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

func (m *mockChecklistRepository) CheckUserHasAccessToChecklist(ctx context.Context, checklistId uint, userId string) (bool, domain.Error) {
	args := m.Called(ctx, checklistId, userId)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(bool), err
}

func (m *mockChecklistRepository) CheckUserIsOwner(ctx context.Context, checklistId uint, userId string) (bool, domain.Error) {
	args := m.Called(ctx, checklistId, userId)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(bool), err
}

func (m *mockChecklistRepository) CreateChecklistShare(ctx context.Context, checklistId uint, sharedByUserId string, sharedWithUserId string) domain.Error {
	args := m.Called(ctx, checklistId, sharedByUserId, sharedWithUserId)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

func (m *mockChecklistRepository) DeleteChecklistShare(ctx context.Context, checklistId uint, userId string) domain.Error {
	args := m.Called(ctx, checklistId, userId)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

// mockChecklistItemsService uses testify's mock for IChecklistItemsService.
type mockChecklistItemsService struct {
	mock.Mock
}

func (m *mockChecklistItemsService) UpdateChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	args := m.Called(ctx, checklistId, checklistItem)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChecklistItem), err
}

func (m *mockChecklistItemsService) SaveChecklistItem(ctx context.Context, checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	args := m.Called(ctx, checklistId, checklistItem)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChecklistItem), err
}

func (m *mockChecklistItemsService) SaveChecklistItemRow(ctx context.Context, checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error) {
	args := m.Called(ctx, checklistId, itemId, row)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChecklistItemRow), err
}

func (m *mockChecklistItemsService) ToggleItemCompleted(ctx context.Context, checklistId uint, checklistItemId uint, completed bool) (domain.ChecklistItem, domain.Error) {
	args := m.Called(ctx, checklistId, checklistItemId, completed)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChecklistItem), err
}

func (m *mockChecklistItemsService) ToggleCompleted(ctx context.Context, checklistId uint, itemId uint, completed bool) (domain.ChecklistItem, domain.Error) {
	args := m.Called(ctx, checklistId, itemId, completed)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChecklistItem), err
}

func (m *mockChecklistItemsService) DeleteChecklistItemRow(ctx context.Context, checklistId uint, itemId uint, rowId uint) domain.Error {
	args := m.Called(ctx, checklistId, itemId, rowId)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

func (m *mockChecklistItemsService) FindChecklistItemById(ctx context.Context, checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
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

func (m *mockChecklistItemsService) DeleteChecklistItemById(ctx context.Context, checklistId uint, id uint) domain.Error {
	args := m.Called(ctx, checklistId, id)
	if arg := args.Get(0); arg != nil {
		return arg.(domain.Error)
	}
	return nil
}

func (m *mockChecklistItemsService) FindAllChecklistItems(ctx context.Context, checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error) {
	args := m.Called(ctx, checklistId, completed, sortOrder)
	var items []domain.ChecklistItem
	var err domain.Error
	if arg := args.Get(0); arg != nil {
		items = arg.([]domain.ChecklistItem)
	}
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return items, err
}

func (m *mockChecklistItemsService) ChangeChecklistItemOrder(ctx context.Context, request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error) {
	args := m.Called(ctx, request)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).(domain.ChangeOrderResponse), err
}

// Test DeleteChecklistById - Success case (empty checklist)
func TestChecklistService_DeleteChecklistById_Success(t *testing.T) {
	ctx := context.Background()
	checklistId := uint(123)

	repo := new(mockChecklistRepository)
	ownershipChecker := new(mockChecklistOwnershipChecker)
	itemService := new(mockChecklistItemsService)

	// Mock expectations
	ownershipChecker.On("HasAccessToChecklist", ctx, checklistId).Return(nil)
	itemService.On("FindAllChecklistItems", ctx, checklistId, (*bool)(nil), domain.AscSort).Return([]domain.ChecklistItem{}, nil)
	repo.On("DeleteChecklistById", ctx, checklistId).Return(nil)

	svc := &checklistService{
		repository:                repo,
		checklistOwnershipChecker: ownershipChecker,
		checklistItemService:      itemService,
	}

	err := svc.DeleteChecklistById(ctx, checklistId)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	ownershipChecker.AssertExpectations(t)
	itemService.AssertExpectations(t)
	repo.AssertExpectations(t)
}

// Test DeleteChecklistById - Access denied (user doesn't have access)
func TestChecklistService_DeleteChecklistById_AccessDenied(t *testing.T) {
	ctx := context.Background()
	checklistId := uint(123)
	expectedErr := domain.NewError("access denied", 403)

	repo := new(mockChecklistRepository)
	ownershipChecker := new(mockChecklistOwnershipChecker)
	itemService := new(mockChecklistItemsService)

	// Mock expectations
	ownershipChecker.On("HasAccessToChecklist", ctx, checklistId).Return(expectedErr)

	svc := &checklistService{
		repository:                repo,
		checklistOwnershipChecker: ownershipChecker,
		checklistItemService:      itemService,
	}

	err := svc.DeleteChecklistById(ctx, checklistId)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "Checklist(id=123) not found" {
		t.Fatalf("expected 'Checklist(id=123) not found', got: %v", err.Error())
	}

	ownershipChecker.AssertExpectations(t)
	// Should not call itemService or repo when access is denied
	itemService.AssertNotCalled(t, "FindAllChecklistItems")
	repo.AssertNotCalled(t, "DeleteChecklistById")
}

// Test DeleteChecklistById - Checklist not empty (has items)
func TestChecklistService_DeleteChecklistById_ChecklistNotEmpty(t *testing.T) {
	ctx := context.Background()
	checklistId := uint(123)

	repo := new(mockChecklistRepository)
	ownershipChecker := new(mockChecklistOwnershipChecker)
	itemService := new(mockChecklistItemsService)

	// Mock expectations - checklist has 2 items
	items := []domain.ChecklistItem{
		{Id: 1, Name: "Item 1", Completed: false},
		{Id: 2, Name: "Item 2", Completed: true},
	}
	ownershipChecker.On("HasAccessToChecklist", ctx, checklistId).Return(nil)
	itemService.On("FindAllChecklistItems", ctx, checklistId, (*bool)(nil), domain.AscSort).Return(items, nil)

	svc := &checklistService{
		repository:                repo,
		checklistOwnershipChecker: ownershipChecker,
		checklistItemService:      itemService,
	}

	err := svc.DeleteChecklistById(ctx, checklistId)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "Checklist is not empty" {
		t.Fatalf("expected 'Checklist is not empty', got: %v", err.Error())
	}
	if err.ResponseCode() != 400 {
		t.Fatalf("expected response code 400, got: %d", err.ResponseCode())
	}

	ownershipChecker.AssertExpectations(t)
	itemService.AssertExpectations(t)
	// Should not call repo.DeleteChecklistById when checklist has items
	repo.AssertNotCalled(t, "DeleteChecklistById")
}

// Test DeleteChecklistById - Error finding items
func TestChecklistService_DeleteChecklistById_FindItemsError(t *testing.T) {
	ctx := context.Background()
	checklistId := uint(123)
	expectedErr := domain.NewError("database error", 500)

	repo := new(mockChecklistRepository)
	ownershipChecker := new(mockChecklistOwnershipChecker)
	itemService := new(mockChecklistItemsService)

	// Mock expectations
	ownershipChecker.On("HasAccessToChecklist", ctx, checklistId).Return(nil)
	itemService.On("FindAllChecklistItems", ctx, checklistId, (*bool)(nil), domain.AscSort).Return(([]domain.ChecklistItem)(nil), expectedErr)

	svc := &checklistService{
		repository:                repo,
		checklistOwnershipChecker: ownershipChecker,
		checklistItemService:      itemService,
	}

	err := svc.DeleteChecklistById(ctx, checklistId)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != expectedErr {
		t.Fatalf("expected error %v, got: %v", expectedErr, err)
	}

	ownershipChecker.AssertExpectations(t)
	itemService.AssertExpectations(t)
	// Should not call repo.DeleteChecklistById when finding items fails
	repo.AssertNotCalled(t, "DeleteChecklistById")
}

// Test DeleteChecklistById - Repository deletion fails
func TestChecklistService_DeleteChecklistById_RepositoryError(t *testing.T) {
	ctx := context.Background()
	checklistId := uint(123)
	expectedErr := domain.NewError("failed to delete", 500)

	repo := new(mockChecklistRepository)
	ownershipChecker := new(mockChecklistOwnershipChecker)
	itemService := new(mockChecklistItemsService)

	// Mock expectations
	ownershipChecker.On("HasAccessToChecklist", ctx, checklistId).Return(nil)
	itemService.On("FindAllChecklistItems", ctx, checklistId, (*bool)(nil), domain.AscSort).Return([]domain.ChecklistItem{}, nil)
	repo.On("DeleteChecklistById", ctx, checklistId).Return(expectedErr)

	svc := &checklistService{
		repository:                repo,
		checklistOwnershipChecker: ownershipChecker,
		checklistItemService:      itemService,
	}

	err := svc.DeleteChecklistById(ctx, checklistId)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != expectedErr {
		t.Fatalf("expected error %v, got: %v", expectedErr, err)
	}

	ownershipChecker.AssertExpectations(t)
	itemService.AssertExpectations(t)
	repo.AssertExpectations(t)
}

// Test DeleteChecklistById - Empty checklist (nil items slice)
func TestChecklistService_DeleteChecklistById_NilItemsSlice(t *testing.T) {
	ctx := context.Background()
	checklistId := uint(123)

	repo := new(mockChecklistRepository)
	ownershipChecker := new(mockChecklistOwnershipChecker)
	itemService := new(mockChecklistItemsService)

	// Mock expectations - nil items slice (treated as empty)
	ownershipChecker.On("HasAccessToChecklist", ctx, checklistId).Return(nil)
	itemService.On("FindAllChecklistItems", ctx, checklistId, (*bool)(nil), domain.AscSort).Return(([]domain.ChecklistItem)(nil), nil)
	repo.On("DeleteChecklistById", ctx, checklistId).Return(nil)

	svc := &checklistService{
		repository:                repo,
		checklistOwnershipChecker: ownershipChecker,
		checklistItemService:      itemService,
	}

	err := svc.DeleteChecklistById(ctx, checklistId)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	ownershipChecker.AssertExpectations(t)
	itemService.AssertExpectations(t)
	repo.AssertExpectations(t)
}
