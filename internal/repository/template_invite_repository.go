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

type templateInviteRepository struct {
	connection pool.Conn
}

func newTemplateInviteRepository(conn pool.Conn) repository.ITemplateInviteRepository {
	return &templateInviteRepository{connection: conn}
}

func (r *templateInviteRepository) CreateInvite(ctx context.Context, invite domain.TemplateInvite) (domain.TemplateInvite, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (domain.TemplateInvite, error) {
		query := `INSERT INTO TEMPLATE_INVITE(ID, TEMPLATE_ID, NAME, INVITE_TOKEN, CREATED_BY, CREATED_AT, EXPIRES_AT, IS_SINGLE_USE)
				  VALUES (nextval('template_invite_id_sequence'), @template_id, @name, @invite_token, @created_by, @created_at, @expires_at, @is_single_use)
				  RETURNING ID`

		row := tx.QueryRow(ctx, query, pgx.NamedArgs{
			"template_id":   invite.TemplateId,
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

	result, err := connection.RunInTransaction(connection.TransactionProps[domain.TemplateInvite]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return domain.TemplateInvite{}, domain.Wrap(err, "Failed to create template invite", 500)
	}

	return result, nil
}

func (r *templateInviteRepository) FindInviteByToken(ctx context.Context, token string) (*domain.TemplateInvite, domain.Error) {
	query := `SELECT id, template_id, name, invite_token, created_by, created_at, expires_at, claimed_by, claimed_at, is_single_use
			  FROM TEMPLATE_INVITE
			  WHERE invite_token = @token`

	var inviteDbo dbo.TemplateInviteDbo
	err := r.connection.QueryOne(ctx, query, &inviteDbo, pgx.NamedArgs{
		"token": token,
	})

	if errors.Is(err, mapper.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, domain.Wrap(err, "Failed to find template invite by token", 500)
	}

	invite := dbo.MapTemplateInviteDboToDomain(inviteDbo)
	return &invite, nil
}

func (r *templateInviteRepository) FindActiveInvitesByTemplateId(ctx context.Context, templateId uint) ([]domain.TemplateInvite, domain.Error) {
	query := `SELECT id, template_id, name, invite_token, created_by, created_at, expires_at, claimed_by, claimed_at, is_single_use
			  FROM TEMPLATE_INVITE
			  WHERE template_id = @template_id
			    AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
			    AND claimed_at IS NULL
			  ORDER BY created_at DESC`

	var inviteDpos []dbo.TemplateInviteDbo
	err := r.connection.QueryList(ctx, query, &inviteDpos, pgx.NamedArgs{
		"template_id": templateId,
	})

	if err != nil {
		return nil, domain.Wrap(err, "Failed to find active template invites", 500)
	}

	invites := make([]domain.TemplateInvite, 0, len(inviteDpos))
	for _, d := range inviteDpos {
		invites = append(invites, dbo.MapTemplateInviteDboToDomain(d))
	}

	return invites, nil
}

func (r *templateInviteRepository) DeleteInviteById(ctx context.Context, inviteId uint) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		query := `DELETE FROM TEMPLATE_INVITE WHERE id = @invite_id`
		result, err := tx.Exec(ctx, query, pgx.NamedArgs{
			"invite_id": inviteId,
		})
		return result.RowsAffected() == 1, err
	}

	success, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return domain.Wrap(err, "Failed to delete template invite", 500)
	}

	if !success {
		return domain.NewError(fmt.Sprintf("Template invite(id=%d) not found", inviteId), 404)
	}

	return nil
}

func (r *templateInviteRepository) ClaimInvite(ctx context.Context, token string, userId string) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		query := `UPDATE TEMPLATE_INVITE
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
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxSerializable,
	})

	if err != nil {
		return domain.Wrap(err, "Failed to claim template invite", 500)
	}

	if !success {
		return domain.NewError("Template invite not found or already claimed", 400)
	}

	return nil
}

func (r *templateInviteRepository) ClaimInviteAndCreateShare(ctx context.Context, token string, userId string, templateId uint, sharedBy string) domain.Error {
	queryFunc := func(tx pool.TransactionWrapper) (bool, error) {
		claimQuery := `UPDATE TEMPLATE_INVITE
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
			return false, fmt.Errorf("template invite not found or already claimed")
		}

		shareQuery := `INSERT INTO TEMPLATE_SHARE(ID, TEMPLATE_ID, SHARED_BY_USER_ID, SHARED_WITH_USER_ID, CREATED_AT)
				  VALUES (nextval('template_share_id_sequence'), @template_id, @shared_by, @shared_with, CURRENT_TIMESTAMP)
				  ON CONFLICT (TEMPLATE_ID, SHARED_WITH_USER_ID) DO NOTHING`

		_, err = tx.Exec(ctx, shareQuery, pgx.NamedArgs{
			"template_id": templateId,
			"shared_by":   sharedBy,
			"shared_with": userId,
		})

		return true, err
	}

	success, err := connection.RunInTransaction(connection.TransactionProps[bool]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxSerializable,
	})

	if err != nil {
		return domain.Wrap(err, "Failed to claim template invite and create share", 500)
	}

	if !success {
		return domain.NewError("Template invite not found or already claimed", 400)
	}

	return nil
}

func (r *templateInviteRepository) DeleteExpiredInvites(ctx context.Context) (int64, domain.Error) {
	queryFunc := func(tx pool.TransactionWrapper) (int64, error) {
		query := `DELETE FROM TEMPLATE_INVITE
				  WHERE expires_at < CURRENT_TIMESTAMP
				    AND (is_single_use = true OR claimed_at IS NOT NULL)`

		result, err := tx.Exec(ctx, query, pgx.NamedArgs{})
		if err != nil {
			return 0, err
		}
		return result.RowsAffected(), nil
	}

	rowsDeleted, err := connection.RunInTransaction(connection.TransactionProps[int64]{
		Ctx:        ctx,
		Query:      queryFunc,
		Connection: r.connection,
		TxOptions:  connection.TxReadCommitted,
	})

	if err != nil {
		return 0, domain.Wrap(err, "Failed to delete expired template invites", 500)
	}

	return rowsDeleted, nil
}

func CreateTemplateInviteRepository(conn pool.Conn) repository.ITemplateInviteRepository {
	return newTemplateInviteRepository(conn)
}
