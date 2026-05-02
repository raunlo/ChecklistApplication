package repository

import (
	"context"
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
	coreRepo "com.raunlo.checklist/internal/core/repository"
	"com.raunlo.checklist/internal/repository/connection"
	"com.raunlo.checklist/internal/repository/dbo"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/raunlo/pgx-with-automapper/mapper"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type workspaceRepository struct {
	connection pool.Conn
}

func (r *workspaceRepository) SaveWorkspace(ctx context.Context, workspace domain.Workspace) (domain.Workspace, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (dbo.WorkspaceDBO, error) {
		var d dbo.WorkspaceDBO
		err := tx.QueryRow(ctx,
			`INSERT INTO workspace(owner_user_id, name, description, is_default, created_at, updated_at)
			 VALUES(@owner, @name, @description, @isDefault, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			 RETURNING id, owner_user_id, name, description, is_default, created_at, updated_at`,
			pgx.NamedArgs{
				"owner":       workspace.OwnerUserId,
				"name":        workspace.Name,
				"description": workspace.Description,
				"isDefault":   workspace.IsDefault,
			}).Scan(&d.Id, &d.OwnerUserId, &d.Name, &d.Description, &d.IsDefault, &d.CreatedAt, &d.UpdatedAt)
		return d, err
	}

	res, err := connection.RunInTransaction(connection.TransactionProps[dbo.WorkspaceDBO]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted,
	})
	if err != nil {
		return domain.Workspace{}, domain.Wrap(err, "Could not save workspace", 500)
	}

	return res.ToDomain(), nil
}

func (r *workspaceRepository) FindWorkspaceById(ctx context.Context, id uint) (*domain.Workspace, domain.Error) {
	userId, _ := domain.GetUserIdFromContext(ctx)

	var d dbo.WorkspaceDBO
	err := r.connection.QueryOne(ctx,
		`SELECT w.id, w.owner_user_id, w.name, w.description, w.is_default, w.created_at, w.updated_at,
		        (w.owner_user_id = @userId) AS is_owner,
		        (SELECT COUNT(*) FROM workspace_member wm WHERE wm.workspace_id = w.id) AS member_count
		 FROM workspace w
		 WHERE w.id = @id`,
		&d,
		pgx.NamedArgs{"id": id, "userId": userId},
	)
	if errors.Is(err, mapper.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, domain.Wrap(err, fmt.Sprintf("Failed to find workspace(id=%d)", id), 500)
	}

	result := d.ToDomain()
	return &result, nil
}

func (r *workspaceRepository) FindWorkspacesByUserId(ctx context.Context, userId string) ([]domain.Workspace, domain.Error) {
	var dbos []dbo.WorkspaceDBO
	err := r.connection.QueryList(ctx,
		`SELECT w.id, w.owner_user_id, w.name, w.description, w.is_default, w.created_at, w.updated_at,
		        (w.owner_user_id = @userId) AS is_owner,
		        (SELECT COUNT(*) FROM workspace_member wm WHERE wm.workspace_id = w.id) AS member_count
		 FROM workspace w
		 JOIN workspace_member wm ON wm.workspace_id = w.id
		 WHERE wm.user_id = @userId
		 ORDER BY w.is_default DESC, w.created_at ASC`,
		&dbos,
		pgx.NamedArgs{"userId": userId},
	)
	if err != nil {
		return nil, domain.Wrap(err, "Failed to find workspaces", 500)
	}

	result := make([]domain.Workspace, 0, len(dbos))
	for _, d := range dbos {
		result = append(result, d.ToDomain())
	}
	return result, nil
}

func (r *workspaceRepository) UpdateWorkspace(ctx context.Context, workspace domain.Workspace) (domain.Workspace, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (dbo.WorkspaceDBO, error) {
		var d dbo.WorkspaceDBO
		err := tx.QueryRow(ctx,
			`UPDATE workspace SET name = @name, description = @description, updated_at = CURRENT_TIMESTAMP
			 WHERE id = @id
			 RETURNING id, owner_user_id, name, description, is_default, created_at, updated_at`,
			pgx.NamedArgs{
				"id":          workspace.Id,
				"name":        workspace.Name,
				"description": workspace.Description,
			}).Scan(&d.Id, &d.OwnerUserId, &d.Name, &d.Description, &d.IsDefault, &d.CreatedAt, &d.UpdatedAt)
		d.IsOwner = true
		return d, err
	}

	res, err := connection.RunInTransaction(connection.TransactionProps[dbo.WorkspaceDBO]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted,
	})
	if err != nil {
		return domain.Workspace{}, domain.Wrap(err, "Could not update workspace", 500)
	}

	return res.ToDomain(), nil
}

func (r *workspaceRepository) DeleteWorkspace(ctx context.Context, id uint) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		result, err := tx.Exec(ctx,
			`DELETE FROM workspace WHERE id = @id`,
			pgx.NamedArgs{"id": id})
		return result.RowsAffected() == 1, err
	}

	_, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted,
	})
	if err != nil {
		return domain.Wrap(err, "Could not delete workspace", 500)
	}
	return nil
}

func (r *workspaceRepository) CheckUserIsWorkspaceOwner(ctx context.Context, workspaceId uint, userId string) (bool, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		var count int
		err := tx.QueryRow(ctx,
			`SELECT COUNT(*) FROM workspace WHERE id = @id AND owner_user_id = @userId`,
			pgx.NamedArgs{"id": workspaceId, "userId": userId}).Scan(&count)
		return count > 0, err
	}

	isOwner, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted,
	})
	if err != nil {
		return false, domain.Wrap(err, "Failed to check workspace ownership", 500)
	}
	return isOwner, nil
}

func (r *workspaceRepository) CheckUserIsMember(ctx context.Context, workspaceId uint, userId string) (bool, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		var count int
		err := tx.QueryRow(ctx,
			`SELECT COUNT(*) FROM workspace_member WHERE workspace_id = @workspaceId AND user_id = @userId`,
			pgx.NamedArgs{"workspaceId": workspaceId, "userId": userId}).Scan(&count)
		return count > 0, err
	}

	isMember, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted,
	})
	if err != nil {
		return false, domain.Wrap(err, "Failed to check workspace membership", 500)
	}
	return isMember, nil
}

func (r *workspaceRepository) GetWorkspaceMembers(ctx context.Context, workspaceId uint) ([]domain.WorkspaceMember, domain.Error) {
	var dbos []dbo.WorkspaceMemberDBO
	err := r.connection.QueryList(ctx,
		`SELECT u.user_id, u.name, u.user_id AS email,
		        (w.owner_user_id = u.user_id) AS is_owner
		 FROM workspace_member wm
		 JOIN app_user u ON u.user_id = wm.user_id
		 JOIN workspace w ON w.id = wm.workspace_id
		 WHERE wm.workspace_id = @workspaceId
		 ORDER BY (w.owner_user_id = u.user_id) DESC, wm.joined_at ASC`,
		&dbos,
		pgx.NamedArgs{"workspaceId": workspaceId},
	)
	if err != nil {
		return nil, domain.Wrap(err, "Failed to get workspace members", 500)
	}

	result := make([]domain.WorkspaceMember, 0, len(dbos))
	for _, d := range dbos {
		result = append(result, d.ToDomain(workspaceId))
	}
	return result, nil
}

func (r *workspaceRepository) RemoveMember(ctx context.Context, workspaceId uint, userId string) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		result, err := tx.Exec(ctx,
			`DELETE FROM workspace_member WHERE workspace_id = @workspaceId AND user_id = @userId`,
			pgx.NamedArgs{"workspaceId": workspaceId, "userId": userId})
		return result.RowsAffected() == 1, err
	}

	success, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted,
	})
	if err != nil {
		return domain.Wrap(err, "Failed to remove workspace member", 500)
	}
	if !success {
		return domain.NewError(fmt.Sprintf("Member not found in workspace(id=%d)", workspaceId), 404)
	}
	return nil
}

func (r *workspaceRepository) AddMember(ctx context.Context, workspaceId uint, userId string) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		_, err := tx.Exec(ctx,
			`INSERT INTO workspace_member(workspace_id, user_id, joined_at)
			 VALUES(@workspaceId, @userId, CURRENT_TIMESTAMP)
			 ON CONFLICT (workspace_id, user_id) DO NOTHING`,
			pgx.NamedArgs{"workspaceId": workspaceId, "userId": userId})
		return true, err
	}

	_, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted,
	})
	if err != nil {
		return domain.Wrap(err, "Failed to add workspace member", 500)
	}
	return nil
}

func (r *workspaceRepository) FindDefaultWorkspace(ctx context.Context, userId string) (*domain.Workspace, domain.Error) {
	var d dbo.WorkspaceDBO
	err := r.connection.QueryOne(ctx,
		`SELECT w.id, w.owner_user_id, w.name, w.description, w.is_default, w.created_at, w.updated_at,
		        TRUE AS is_owner,
		        (SELECT COUNT(*) FROM workspace_member wm WHERE wm.workspace_id = w.id) AS member_count
		 FROM workspace w
		 WHERE w.owner_user_id = @userId AND w.is_default = TRUE
		 LIMIT 1`,
		&d,
		pgx.NamedArgs{"userId": userId},
	)
	if errors.Is(err, mapper.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, domain.Wrap(err, "Failed to find default workspace", 500)
	}
	result := d.ToDomain()
	return &result, nil
}

func CreateWorkspaceRepository(connection pool.Conn) coreRepo.IWorkspaceRepository {
	return &workspaceRepository{connection: connection}
}
