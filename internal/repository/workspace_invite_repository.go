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

type workspaceInviteRepository struct {
	connection pool.Conn
}

func (r *workspaceInviteRepository) CreateInvite(ctx context.Context, invite domain.WorkspaceInvite) (domain.WorkspaceInvite, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (domain.WorkspaceInvite, error) {
		row := tx.QueryRow(ctx,
			`INSERT INTO workspace_invite(workspace_id, name, invite_token, created_by, created_at, expires_at, is_single_use)
			 VALUES(@workspace_id, @name, @invite_token, @created_by, @created_at, @expires_at, @is_single_use)
			 RETURNING id`,
			pgx.NamedArgs{
				"workspace_id":  invite.WorkspaceId,
				"name":          invite.Name,
				"invite_token":  invite.InviteToken,
				"created_by":    invite.CreatedBy,
				"created_at":    invite.CreatedAt,
				"expires_at":    invite.ExpiresAt,
				"is_single_use": invite.IsSingleUse,
			})
		err := row.Scan(&invite.Id)
		return invite, err
	}

	result, err := connection.RunInTransaction(connection.TransactionProps[domain.WorkspaceInvite]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted,
	})
	if err != nil {
		return domain.WorkspaceInvite{}, domain.Wrap(err, "Failed to create workspace invite", 500)
	}
	return result, nil
}

func (r *workspaceInviteRepository) FindInviteByToken(ctx context.Context, token string) (*domain.WorkspaceInvite, domain.Error) {
	var d dbo.WorkspaceInviteDBO
	err := r.connection.QueryOne(ctx,
		`SELECT id, workspace_id, name, invite_token, created_by, created_at, expires_at, claimed_by, claimed_at, is_single_use
		 FROM workspace_invite WHERE invite_token = @token`,
		&d,
		pgx.NamedArgs{"token": token},
	)
	if errors.Is(err, mapper.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, domain.Wrap(err, "Failed to find workspace invite by token", 500)
	}
	invite := dbo.MapWorkspaceInviteDboToDomain(d)
	return &invite, nil
}

func (r *workspaceInviteRepository) FindActiveInvitesByWorkspaceId(ctx context.Context, workspaceId uint) ([]domain.WorkspaceInvite, domain.Error) {
	var dbos []dbo.WorkspaceInviteDBO
	err := r.connection.QueryList(ctx,
		`SELECT id, workspace_id, name, invite_token, created_by, created_at, expires_at, claimed_by, claimed_at, is_single_use
		 FROM workspace_invite
		 WHERE workspace_id = @workspaceId
		   AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		   AND (is_single_use = FALSE OR claimed_at IS NULL)
		 ORDER BY created_at DESC`,
		&dbos,
		pgx.NamedArgs{"workspaceId": workspaceId},
	)
	if err != nil {
		return nil, domain.Wrap(err, "Failed to find active workspace invites", 500)
	}

	result := make([]domain.WorkspaceInvite, 0, len(dbos))
	for _, d := range dbos {
		result = append(result, dbo.MapWorkspaceInviteDboToDomain(d))
	}
	return result, nil
}

func (r *workspaceInviteRepository) DeleteInviteById(ctx context.Context, inviteId uint) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		result, err := tx.Exec(ctx,
			`DELETE FROM workspace_invite WHERE id = @inviteId`,
			pgx.NamedArgs{"inviteId": inviteId})
		return result.RowsAffected() == 1, err
	}

	success, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted,
	})
	if err != nil {
		return domain.Wrap(err, "Failed to delete workspace invite", 500)
	}
	if !success {
		return domain.NewError(fmt.Sprintf("Workspace invite(id=%d) not found", inviteId), 404)
	}
	return nil
}

func (r *workspaceInviteRepository) ClaimInviteAndAddMember(ctx context.Context, token string, userId string, workspaceId uint, isSingleUse bool) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		if isSingleUse {
			claimResult, err := tx.Exec(ctx,
				`UPDATE workspace_invite
				 SET claimed_by = @userId, claimed_at = CURRENT_TIMESTAMP
				 WHERE invite_token = @token AND claimed_at IS NULL`,
				pgx.NamedArgs{"token": token, "userId": userId})
			if err != nil {
				return false, err
			}
			if claimResult.RowsAffected() != 1 {
				return false, fmt.Errorf("workspace invite not found or already claimed")
			}
		}

		_, err := tx.Exec(ctx,
			`INSERT INTO workspace_member(workspace_id, user_id, joined_at)
			 VALUES(@workspaceId, @userId, CURRENT_TIMESTAMP)
			 ON CONFLICT (workspace_id, user_id) DO NOTHING`,
			pgx.NamedArgs{"workspaceId": workspaceId, "userId": userId})
		return true, err
	}

	success, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxSerializable,
	})
	if err != nil {
		return domain.Wrap(err, "Failed to claim workspace invite", 500)
	}
	if !success {
		return domain.NewError("Workspace invite not found or already claimed", 400)
	}
	return nil
}

func CreateWorkspaceInviteRepository(conn pool.Conn) coreRepo.IWorkspaceInviteRepository {
	return &workspaceInviteRepository{connection: conn}
}
