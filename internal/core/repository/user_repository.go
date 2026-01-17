package repository

import (
	"context"
	"time"

	"com.raunlo.checklist/internal/core/domain"
)

type IUserRepository interface {
	DeleteAllUserChecklists(ctx context.Context, userId string) error
	GetUserDataExport(ctx context.Context, userId string) (*domain.UserDataExport, error)
	CreateOrUpdateUser(ctx context.Context, user domain.User) domain.Error
	FindUserById(ctx context.Context, userId string) (*domain.User, domain.Error)
}

type ExportedChecklist struct {
	Id        uint
	Name      string
	CreatedAt time.Time
	Items     []ExportedChecklistItem
	Shares    []ExportedChecklistShare
}

type ExportedChecklistItem struct {
	Id          uint
	Name        string
	Completed   bool
	OrderNumber int
	Rows        []ExportedChecklistItemRow
}

type ExportedChecklistItemRow struct {
	Id        uint
	Name      string
	Completed bool
}

type ExportedChecklistShare struct {
	SharedWithUserId string
	PermissionLevel  string
	SharedAt         time.Time
}
