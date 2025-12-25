package guardrail

import (
	"context"
	"testing"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

// mockChecklistRepository is a mock implementation of repository.IChecklistRepository
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
	if arg := args.Get(0); arg != nil {
		checklist = arg.(*domain.Checklist)
	}
	var err domain.Error
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

func (m *mockChecklistRepository) CheckUserHasAccessToChecklist(ctx context.Context, checklistId uint, userId string) (bool, domain.Error) {
	args := m.Called(ctx, checklistId, userId)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Bool(0), err
}

func (m *mockChecklistRepository) CheckUserIsOwner(ctx context.Context, checklistId uint, userId string) (bool, domain.Error) {
	args := m.Called(ctx, checklistId, userId)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Bool(0), err
}

func (m *mockChecklistRepository) FindAllChecklists(ctx context.Context) ([]domain.Checklist, domain.Error) {
	args := m.Called(ctx)
	var err domain.Error
	if arg := args.Get(1); arg != nil {
		err = arg.(domain.Error)
	}
	return args.Get(0).([]domain.Checklist), err
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

// TestHasAccessToChecklist_ValidOwnerAccess tests that an owner can access their checklist
func TestHasAccessToChecklist_ValidOwnerAccess(t *testing.T) {
	repo := new(mockChecklistRepository)
	service := NewChecklistOwnershipCheckerService(repo)

	userId := "user-123"
	checklistId := uint(1)

	ctx := context.WithValue(context.Background(), domain.UserIdContextKey, userId)
	repo.On("CheckUserHasAccessToChecklist", mock.Anything, checklistId, userId).Return(true, nil)

	err := service.HasAccessToChecklist(ctx, checklistId)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	repo.AssertExpectations(t)
}

// TestHasAccessToChecklist_ValidSharedAccess tests that a user with shared access can access the checklist
func TestHasAccessToChecklist_ValidSharedAccess(t *testing.T) {
	repo := new(mockChecklistRepository)
	service := NewChecklistOwnershipCheckerService(repo)

	sharedUserId := "shared-user-456"
	checklistId := uint(2)

	ctx := context.WithValue(context.Background(), domain.UserIdContextKey, sharedUserId)
	repo.On("CheckUserHasAccessToChecklist", mock.Anything, checklistId, sharedUserId).Return(true, nil)

	err := service.HasAccessToChecklist(ctx, checklistId)
	if err != nil {
		t.Fatalf("expected no error for shared access, got: %v", err)
	}
	repo.AssertExpectations(t)
}

// TestHasAccessToChecklist_AccessDeniedForUnauthorizedUser tests that unauthorized users are denied access
func TestHasAccessToChecklist_AccessDeniedForUnauthorizedUser(t *testing.T) {
	repo := new(mockChecklistRepository)
	service := NewChecklistOwnershipCheckerService(repo)

	unauthorizedUserId := "unauthorized-user-789"
	checklistId := uint(3)

	ctx := context.WithValue(context.Background(), domain.UserIdContextKey, unauthorizedUserId)
	repo.On("CheckUserHasAccessToChecklist", mock.Anything, checklistId, unauthorizedUserId).Return(false, nil)

	err := service.HasAccessToChecklist(ctx, checklistId)
	if err == nil {
		t.Fatalf("expected error for unauthorized user, got nil")
	}
	if err.ResponseCode() != 404 {
		t.Fatalf("expected 404 response code, got: %d", err.ResponseCode())
	}
	repo.AssertExpectations(t)
}

// TestHasAccessToChecklist_MissingUserIdInContext tests handling when user ID is not in context
func TestHasAccessToChecklist_MissingUserIdInContext(t *testing.T) {
	repo := new(mockChecklistRepository)
	service := NewChecklistOwnershipCheckerService(repo)

	checklistId := uint(4)
	ctx := context.Background() // No user ID in context

	err := service.HasAccessToChecklist(ctx, checklistId)
	if err == nil {
		t.Fatalf("expected error for missing user ID, got nil")
	}
	if err.ResponseCode() != 401 {
		t.Fatalf("expected 401 response code, got: %d", err.ResponseCode())
	}
	// Repository should not be called when user ID is missing
	repo.AssertNotCalled(t, "CheckUserHasAccessToChecklist", mock.Anything, mock.Anything)
}

// TestHasAccessToChecklist_EmptyUserIdInContext tests handling when user ID is empty string
func TestHasAccessToChecklist_EmptyUserIdInContext(t *testing.T) {
	repo := new(mockChecklistRepository)
	service := NewChecklistOwnershipCheckerService(repo)

	checklistId := uint(5)
	ctx := context.WithValue(context.Background(), domain.UserIdContextKey, "")

	err := service.HasAccessToChecklist(ctx, checklistId)
	if err == nil {
		t.Fatalf("expected error for empty user ID, got nil")
	}
	if err.ResponseCode() != 401 {
		t.Fatalf("expected 401 response code, got: %d", err.ResponseCode())
	}
	// Repository should not be called when user ID is empty
	repo.AssertNotCalled(t, "CheckUserHasAccessToChecklist", mock.Anything, mock.Anything)
}

// TestHasAccessToChecklist_InvalidUserIdTypeInContext tests handling when user ID has wrong type
func TestHasAccessToChecklist_InvalidUserIdTypeInContext(t *testing.T) {
	repo := new(mockChecklistRepository)
	service := NewChecklistOwnershipCheckerService(repo)

	checklistId := uint(6)
	// Store an int instead of string
	ctx := context.WithValue(context.Background(), domain.UserIdContextKey, 12345)

	err := service.HasAccessToChecklist(ctx, checklistId)
	if err == nil {
		t.Fatalf("expected error for invalid user ID type, got nil")
	}
	if err.ResponseCode() != 401 {
		t.Fatalf("expected 401 response code, got: %d", err.ResponseCode())
	}
	// Repository should not be called when user ID type is invalid
	repo.AssertNotCalled(t, "CheckUserHasAccessToChecklist", mock.Anything, mock.Anything)
}

// TestHasAccessToChecklist_RepositoryError tests handling of repository errors
func TestHasAccessToChecklist_RepositoryError(t *testing.T) {
	repo := new(mockChecklistRepository)
	service := NewChecklistOwnershipCheckerService(repo)

	userId := "user-123"
	checklistId := uint(7)
	expectedErr := domain.NewError("database connection error", 500)

	ctx := context.WithValue(context.Background(), domain.UserIdContextKey, userId)
	repo.On("CheckUserHasAccessToChecklist", mock.Anything, checklistId, userId).Return(false, expectedErr)

	err := service.HasAccessToChecklist(ctx, checklistId)
	if err == nil {
		t.Fatalf("expected error from repository, got nil")
	}
	if err.ResponseCode() != 500 {
		t.Fatalf("expected 500 response code, got: %d", err.ResponseCode())
	}
	repo.AssertExpectations(t)
}

// TestHasAccessToChecklist_DifferentChecklistIds tests access check with different checklist IDs
func TestHasAccessToChecklist_DifferentChecklistIds(t *testing.T) {
	repo := new(mockChecklistRepository)
	service := NewChecklistOwnershipCheckerService(repo)

	userId := "user-123"

	testCases := []struct {
		checklistId uint
		hasAccess   bool
	}{
		{checklistId: 1, hasAccess: true},
		{checklistId: 2, hasAccess: false},
		{checklistId: 3, hasAccess: true},
	}

	for _, tc := range testCases {
		ctx := context.WithValue(context.Background(), domain.UserIdContextKey, userId)
		repo.On("CheckUserHasAccessToChecklist", mock.Anything, tc.checklistId, userId).Return(tc.hasAccess, nil)

		err := service.HasAccessToChecklist(ctx, tc.checklistId)

		if tc.hasAccess {
			if err != nil {
				t.Fatalf("checklist %d: expected no error, got: %v", tc.checklistId, err)
			}
		} else {
			if err == nil {
				t.Fatalf("checklist %d: expected error for no access, got nil", tc.checklistId)
			}
		}
	}
	repo.AssertExpectations(t)
}
