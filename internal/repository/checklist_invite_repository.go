package repository

import (
	"context"
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/repository"
	"com.raunlo.checklist/internal/repository/connection"
	"com.raunlo.checklist/internal/repository/dbo"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/raunlo/pgx-with-automapper/mapper"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type checklistInviteRepository struct {
	connection pool.Conn
}

func newChecklistInviteRepository(connection pool.Conn) repository.IChecklistInviteRepository {
	return &checklistInviteRepository{
		connection: connection,
	}
}

func (r *checklistInviteRepository) CreateInvite(ctx context.Context, invite domain.ChecklistInvite) (domain.ChecklistInvite, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (domain.ChecklistInvite, error) {
		query := `INSERT INTO CHECKLIST_INVITE(ID, CHECKLIST_ID, NAME, INVITE_TOKEN, CREATED_BY, CREATED_AT, EXPIRES_AT, IS_SINGLE_USE)
				  VALUES (nextval('checklist_invite_id_sequence'), @checklist_id, @name, @invite_token, @created_by, @created_at, @expires_at, @is_single_use)
				  RETURNING ID`

		row := tx.QueryRow(ctx, query, pgx.NamedArgs{
			"checklist_id":  invite.ChecklistId,
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

	result, err := connection.RunInTransaction(connection.TransactionProps[domain.ChecklistInvite]{
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted, // Simple single-row insert
	})

	if err != nil {
		return domain.ChecklistInvite{}, domain.Wrap(err, "Failed to create invite", 500)
	}

	return result, nil
}

func (r *checklistInviteRepository) FindInviteByToken(ctx context.Context, token string) (*domain.ChecklistInvite, domain.Error) {
	query := `SELECT id, checklist_id, name, invite_token, created_by, created_at, expires_at, claimed_by, claimed_at, is_single_use
			  FROM CHECKLIST_INVITE
			  WHERE invite_token = @token`

	var inviteDbo dbo.ChecklistInviteDbo
	err := r.connection.QueryOne(ctx, query, &inviteDbo, pgx.NamedArgs{
		"token": token,
	})

	if errors.Is(err, mapper.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, domain.Wrap(err, "Failed to find invite by token", 500)
	}

	invite := dbo.MapChecklistInviteDboToDomain(inviteDbo)
	return &invite, nil
}

func (r *checklistInviteRepository) FindActiveInvitesByChecklistId(ctx context.Context, checklistId uint) ([]domain.ChecklistInvite, domain.Error) {
	query := `SELECT id, checklist_id, name, invite_token, created_by, created_at, expires_at, claimed_by, claimed_at, is_single_use
			  FROM CHECKLIST_INVITE
			  WHERE checklist_id = @checklist_id
			    AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
			    AND claimed_at IS NULL
			  ORDER BY created_at DESC`

	var inviteDpos []dbo.ChecklistInviteDbo
	err := r.connection.QueryList(ctx, query, &inviteDpos, pgx.NamedArgs{
		"checklist_id": checklistId,
	})

	if err != nil {
		return nil, domain.Wrap(err, "Failed to find active invites", 500)
	}

	invites := make([]domain.ChecklistInvite, 0, len(inviteDpos))
	for _, inviteDbo := range inviteDpos {
		invites = append(invites, dbo.MapChecklistInviteDboToDomain(inviteDbo))
	}

	return invites, nil
}

func (r *checklistInviteRepository) DeleteInviteById(ctx context.Context, inviteId uint) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		query := `DELETE FROM CHECKLIST_INVITE WHERE id = @invite_id`
		result, err := tx.Exec(ctx, query, pgx.NamedArgs{
			"invite_id": inviteId,
		})
		return result.RowsAffected() == 1, err
	}

	success, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted, // Simple single-row delete
	})

	if err != nil {
		return domain.Wrap(err, "Failed to delete invite", 500)
	}

	if !success {
		return domain.NewError(fmt.Sprintf("Invite(id=%d) not found", inviteId), 404)
	}

	return nil
}

func (r *checklistInviteRepository) ClaimInvite(ctx context.Context, token string, userId string) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		query := `UPDATE CHECKLIST_INVITE
				  SET claimed_by = @user_id, claimed_at = CURRENT_TIMESTAMP
				  WHERE invite_token = @token
				    AND claimed_at IS NULL`

		result, err := tx.Exec(ctx, query, pgx.NamedArgs{
			"token":   token,
			"user_id": userId,
		})

		return result.RowsAffected() == 1, err
	}

	success, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxSerializable, // Prevents double-claiming race condition
	})

	if err != nil {
		return domain.Wrap(err, "Failed to claim invite", 500)
	}

	if !success {
		return domain.NewError("Invite not found or already claimed", 400)
	}

	return nil
}

// ClaimInviteAndCreateShare atomically claims an invite and creates the share in a single transaction
func (r *checklistInviteRepository) ClaimInviteAndCreateShare(ctx context.Context, token string, userId string, checklistId uint, sharedBy string) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		// First, claim the invite
		claimQuery := `UPDATE CHECKLIST_INVITE
				  SET claimed_by = @user_id, claimed_at = CURRENT_TIMESTAMP
				  WHERE invite_token = @token
				    AND claimed_at IS NULL`

		claimResult, err := tx.Exec(ctx, claimQuery, pgx.NamedArgs{
			"token":   token,
			"user_id": userId,
		})
		if err != nil {
			return false, err
		}

		if claimResult.RowsAffected() != 1 {
			return false, fmt.Errorf("invite not found or already claimed")
		}

		// Then, create the share
		shareQuery := `INSERT INTO CHECKLIST_SHARE(ID, CHECKLIST_ID, SHARED_BY_USER_ID, SHARED_WITH_USER_ID, PERMISSION_LEVEL, CREATED_AT)
				  VALUES (nextval('checklist_share_id_sequence'), @checklist_id, @shared_by, @shared_with, @permission_level, CURRENT_TIMESTAMP)
				  ON CONFLICT (CHECKLIST_ID, SHARED_WITH_USER_ID) DO NOTHING`

		_, err = tx.Exec(ctx, shareQuery, pgx.NamedArgs{
			"checklist_id":     checklistId,
			"shared_by":        sharedBy,
			"shared_with":      userId,
			"permission_level": "READ",
		})

		return true, err
	}

	success, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxSerializable, // Atomic multi-operation: claim + share creation
	})

	if err != nil {
		return domain.Wrap(err, "Failed to claim invite and create share", 500)
	}

	if !success {
		return domain.NewError("Invite not found or already claimed", 400)
	}

	return nil
}

func (r *checklistInviteRepository) DeleteExpiredInvites(ctx context.Context) (int64, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (int64, error) {
		query := `DELETE FROM CHECKLIST_INVITE
				  WHERE expires_at < CURRENT_TIMESTAMP
				    AND (is_single_use = true OR claimed_at IS NOT NULL)`

		result, err := tx.Exec(ctx, query, pgx.NamedArgs{})
		if err != nil {
			return 0, err
		}
		return result.RowsAffected(), nil
	}

	rowsDeleted, err := connection.RunInTransaction(connection.TransactionProps[int64]{
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted, // Cleanup operation, eventual consistency OK
	})

	if err != nil {
		return 0, domain.Wrap(err, "Failed to delete expired invites", 500)
	}

	return rowsDeleted, nil
}
