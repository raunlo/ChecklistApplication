package service

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/repository"
)

type IUserService interface {
	DeleteAccount(ctx context.Context, userId string) error
	ExportUserData(ctx context.Context, userId string) (*domain.UserDataExport, error)
	CreateOrUpdateUser(ctx context.Context, user domain.User) domain.Error
	GetUserById(ctx context.Context, userId string) (*domain.User, domain.Error)
}

type userServiceImpl struct {
	userRepository repository.IUserRepository
}

func NewUserService(userRepository repository.IUserRepository) IUserService {
	return &userServiceImpl{
		userRepository: userRepository,
	}
}

func (s *userServiceImpl) DeleteAccount(ctx context.Context, userId string) error {
	// Delete all checklists owned by the user (CASCADE will handle related data)
	return s.userRepository.DeleteAllUserChecklists(ctx, userId)
}

func (s *userServiceImpl) ExportUserData(ctx context.Context, userId string) (*domain.UserDataExport, error) {
	return s.userRepository.GetUserDataExport(ctx, userId)
}

func (s *userServiceImpl) CreateOrUpdateUser(ctx context.Context, user domain.User) domain.Error {
	return s.userRepository.CreateOrUpdateUser(ctx, user)
}

func (s *userServiceImpl) GetUserById(ctx context.Context, userId string) (*domain.User, domain.Error) {
	return s.userRepository.FindUserById(ctx, userId)
}
