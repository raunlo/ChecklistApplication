package service

import (
	"testing"

	"com.raunlo.checklist/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

// mockChecklistItemsRepository uses testify's mock for repository.IChecklistItemsRepository.
type mockChecklistItemsRepository struct {
	mock.Mock
}

func (m *mockChecklistItemsRepository) UpdateChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	return domain.ChecklistItem{}, nil
}

func (m *mockChecklistItemsRepository) SaveChecklistItem(checklistId uint, checklistItem domain.ChecklistItem) (domain.ChecklistItem, domain.Error) {
	return domain.ChecklistItem{}, nil
}

func (m *mockChecklistItemsRepository) SaveChecklistItemRow(checklistId uint, itemId uint, row domain.ChecklistItemRow) (domain.ChecklistItemRow, domain.Error) {
    args := m.Called(checklistId, itemId, row)
    var err domain.Error
    if e, ok := args.Get(1).(domain.Error); ok {
        err = e
    }
    return args.Get(0).(domain.ChecklistItemRow), err
}

func (m *mockChecklistItemsRepository) FindChecklistItemById(checklistId uint, id uint) (*domain.ChecklistItem, domain.Error) {
	return nil, nil
}

func (m *mockChecklistItemsRepository) DeleteChecklistItemById(checklistId uint, id uint) domain.Error {
	return nil
}

func (m *mockChecklistItemsRepository) FindAllChecklistItems(checklistId uint, completed *bool, sortOrder domain.SortOrder) ([]domain.ChecklistItem, domain.Error) {
	return nil, nil
}

func (m *mockChecklistItemsRepository) ChangeChecklistItemOrder(request domain.ChangeOrderRequest) (domain.ChangeOrderResponse, domain.Error) {
	return domain.ChangeOrderResponse{}, nil
}

func TestChecklistItemsService_SaveChecklistItemRow(t *testing.T) {
	expected := domain.ChecklistItemRow{Id: 1, Name: "row", Completed: false}
	repo := new(mockChecklistItemsRepository)
	repo.On("SaveChecklistItemRow", uint(10), uint(20), domain.ChecklistItemRow{Name: "row"}).Return(expected, nil)

	svc := &checklistItemsService{repository: repo}
	row, err := svc.SaveChecklistItemRow(10, 20, domain.ChecklistItemRow{Name: "row"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row != expected {
		t.Fatalf("expected %#v got %#v", expected, row)
	}
	repo.AssertExpectations(t)
}

func TestChecklistItemsService_SaveChecklistItemRow_Error(t *testing.T) {
	expectedErr := domain.NewError("fail", 500)
	repo := new(mockChecklistItemsRepository)
	repo.On("SaveChecklistItemRow", uint(1), uint(2), domain.ChecklistItemRow{Name: "x"}).Return(domain.ChecklistItemRow{}, expectedErr)

	svc := &checklistItemsService{repository: repo}
	_, err := svc.SaveChecklistItemRow(1, 2, domain.ChecklistItemRow{Name: "x"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != expectedErr {
		t.Fatalf("expected error %v got %v", expectedErr, err)
	}
	repo.AssertExpectations(t)
}
