package repository

import (
	"context"
	"time"

	"com.raunlo.checklist/internal/core/domain"
	coreRepo "com.raunlo.checklist/internal/core/repository"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/raunlo/pgx-with-automapper/pool"
)

type sessionRepository struct {
	connection pool.Conn
}

func NewSessionRepository(connection pool.Conn) coreRepo.ISessionRepository {
	return &sessionRepository{connection: connection}
}

func (r *sessionRepository) CreateSession(ctx context.Context, session domain.Session) (domain.Session, domain.Error) {
	query := `
		INSERT INTO user_session(
			session_id, user_id, access_token_encrypted, refresh_token_encrypted,
			access_token_expires_at, created_at, last_activity_at, expires_at,
			is_invalidated
		) VALUES (
			@session_id, @user_id, @access_token_enc, @refresh_token_enc,
			@access_token_exp, @created_at, @last_activity, @expires_at,
			false
		) RETURNING id`

	err := r.connection.QueryRow(ctx, query, pgx.NamedArgs{
		"session_id":        session.SessionId,
		"user_id":           session.UserId,
		"access_token_enc":  session.AccessTokenEncrypted,
		"refresh_token_enc": session.RefreshTokenEncrypted,
		"access_token_exp":  session.AccessTokenExpiresAt,
		"created_at":        session.CreatedAt,
		"last_activity":     session.LastActivityAt,
		"expires_at":        session.ExpiresAt,
	}).Scan(&session.Id)

	if err != nil {
		return domain.Session{}, domain.Wrap(err, "Failed to create session", 500)
	}

	return session, nil
}

func (r *sessionRepository) FindSessionBySessionId(ctx context.Context, sessionId string) (*domain.Session, domain.Error) {
	query := `
		SELECT id, session_id, user_id, access_token_encrypted, refresh_token_encrypted,
			   access_token_expires_at, created_at, last_activity_at, expires_at,
			   is_invalidated, invalidated_at, invalidation_reason
		FROM user_session
		WHERE session_id = @session_id AND is_invalidated = FALSE`

	var session domain.Session
	err := r.connection.QueryRow(ctx, query, pgx.NamedArgs{"session_id": sessionId}).Scan(
		&session.Id, &session.SessionId, &session.UserId,
		&session.AccessTokenEncrypted, &session.RefreshTokenEncrypted,
		&session.AccessTokenExpiresAt, &session.CreatedAt, &session.LastActivityAt,
		&session.ExpiresAt, &session.IsInvalidated, &session.InvalidatedAt,
		&session.InvalidationReason,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, domain.Wrap(err, "Failed to find session", 500)
	}

	return &session, nil
}

func (r *sessionRepository) UpdateSessionActivity(ctx context.Context, sessionId string, lastActivityAt time.Time) domain.Error {
	query := `
		UPDATE user_session
		SET last_activity_at = @last_activity
		WHERE session_id = @session_id AND is_invalidated = FALSE`

	result, err := r.connection.Exec(ctx, query, pgx.NamedArgs{
		"session_id":    sessionId,
		"last_activity": lastActivityAt,
	})

	if err != nil {
		return domain.Wrap(err, "Failed to update session activity", 500)
	}

	if result.RowsAffected() == 0 {
		return domain.NewError("Session not found or already invalidated", 404)
	}

	return nil
}

func (r *sessionRepository) UpdateSessionTokens(ctx context.Context, sessionId string, accessTokenEnc, refreshTokenEnc []byte, accessTokenExpiresAt time.Time) domain.Error {
	query := `
		UPDATE user_session
		SET access_token_encrypted = @access_token_enc,
			refresh_token_encrypted = @refresh_token_enc,
			access_token_expires_at = @access_token_exp
		WHERE session_id = @session_id AND is_invalidated = FALSE`

	result, err := r.connection.Exec(ctx, query, pgx.NamedArgs{
		"session_id":       sessionId,
		"access_token_enc": accessTokenEnc,
		"refresh_token_enc": refreshTokenEnc,
		"access_token_exp": accessTokenExpiresAt,
	})

	if err != nil {
		return domain.Wrap(err, "Failed to update session tokens", 500)
	}

	if result.RowsAffected() == 0 {
		return domain.NewError("Session not found or already invalidated", 404)
	}

	return nil
}

func (r *sessionRepository) InvalidateSession(ctx context.Context, sessionId string, reason string) domain.Error {
	query := `
		UPDATE user_session
		SET is_invalidated = true,
			invalidated_at = @invalidated_at,
			invalidation_reason = @reason
		WHERE session_id = @session_id`

	_, err := r.connection.Exec(ctx, query, pgx.NamedArgs{
		"session_id":    sessionId,
		"invalidated_at": time.Now(),
		"reason":        reason,
	})

	if err != nil {
		return domain.Wrap(err, "Failed to invalidate session", 500)
	}

	return nil
}

func (r *sessionRepository) InvalidateAllUserSessions(ctx context.Context, userId string, reason string) domain.Error {
	query := `
		UPDATE user_session
		SET is_invalidated = true,
			invalidated_at = @invalidated_at,
			invalidation_reason = @reason
		WHERE user_id = @user_id AND is_invalidated = FALSE`

	_, err := r.connection.Exec(ctx, query, pgx.NamedArgs{
		"user_id":        userId,
		"invalidated_at": time.Now(),
		"reason":         reason,
	})

	if err != nil {
		return domain.Wrap(err, "Failed to invalidate user sessions", 500)
	}

	return nil
}

func (r *sessionRepository) DeleteExpiredSessions(ctx context.Context) (int64, domain.Error) {
	query := `
		DELETE FROM user_session
		WHERE (expires_at < @now OR is_invalidated = TRUE)
		  AND last_activity_at < @cutoff`

	now := time.Now()
	cutoff := now.Add(-7 * 24 * time.Hour) // Keep sessions for 7 days for debugging

	result, err := r.connection.Exec(ctx, query, pgx.NamedArgs{
		"now":    now,
		"cutoff": cutoff,
	})

	if err != nil {
		return 0, domain.Wrap(err, "Failed to delete expired sessions", 500)
	}

	return result.RowsAffected(), nil
}

func (r *sessionRepository) FindSessionsNeedingTokenRefresh(ctx context.Context, beforeTime time.Time) ([]domain.Session, domain.Error) {
	query := `
		SELECT id, session_id, user_id, access_token_encrypted, refresh_token_encrypted,
			   access_token_expires_at, created_at, last_activity_at, expires_at,
			   is_invalidated, invalidated_at, invalidation_reason
		FROM user_session
		WHERE is_invalidated = FALSE
		  AND access_token_expires_at < @before_time
		  AND expires_at > @now`

	rows, err := r.connection.Query(ctx, query, pgx.NamedArgs{
		"before_time": beforeTime,
		"now":         time.Now(),
	})
	if err != nil {
		return nil, domain.Wrap(err, "Failed to find sessions needing refresh", 500)
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var session domain.Session
		err := rows.Scan(
			&session.Id, &session.SessionId, &session.UserId,
			&session.AccessTokenEncrypted, &session.RefreshTokenEncrypted,
			&session.AccessTokenExpiresAt, &session.CreatedAt, &session.LastActivityAt,
			&session.ExpiresAt, &session.IsInvalidated, &session.InvalidatedAt,
			&session.InvalidationReason,
		)
		if err != nil {
			return nil, domain.Wrap(err, "Failed to scan session row", 500)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.Wrap(err, "Error iterating session rows", 500)
	}

	return sessions, nil
}
