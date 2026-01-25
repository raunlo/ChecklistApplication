package repository

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/repository"
	"com.raunlo.checklist/internal/repository/query"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type userRepositoryImpl struct {
	connection pool.Conn
}

func NewUserRepository(connection pool.Conn) repository.IUserRepository {
	return &userRepositoryImpl{
		connection: connection,
	}
}

func (repo *userRepositoryImpl) DeleteAllUserChecklists(ctx context.Context, userId string) error {
	// Delete all checklists owned by the user
	// CASCADE will automatically delete all related data (items, rows, shares, invites)
	_, err := repo.connection.Exec(ctx, query.DeleteAllUserChecklists, userId)
	return err
}

func (repo *userRepositoryImpl) GetUserDataExport(ctx context.Context, userId string) (*domain.UserDataExport, error) {
	return query.GetUserDataExport(ctx, repo.connection, userId)
}

func (repo *userRepositoryImpl) CreateOrUpdateUser(ctx context.Context, user domain.User) domain.Error {
	// Audit timestamps (created_at, updated_at) are handled by SQL DEFAULT CURRENT_TIMESTAMP
	query := `
		INSERT INTO app_user (user_id, name)
		VALUES ($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = CURRENT_TIMESTAMP`

	_, err := repo.connection.Exec(ctx, query, user.UserId, user.Name)

	if err != nil {
		return domain.Wrap(err, "Failed to create or update user", 500)
	}

	return nil
}

func (repo *userRepositoryImpl) FindUserById(ctx context.Context, userId string) (*domain.User, domain.Error) {
	query := `
		SELECT user_id, name
		FROM app_user
		WHERE user_id = $1`

	var user domain.User
	err := repo.connection.QueryRow(ctx, query, userId).Scan(
		&user.UserId, &user.Name,
	)

	if err != nil {
		return nil, domain.Wrap(err, "Failed to find user", 500)
	}

	return &user, nil
}
