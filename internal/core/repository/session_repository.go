package repository

import (
	"context"
	"time"

	"com.raunlo.checklist/internal/core/domain"
)

type ISessionRepository interface {
	CreateSession(ctx context.Context, session domain.Session) (domain.Session, domain.Error)
	FindSessionBySessionId(ctx context.Context, sessionId string) (*domain.Session, domain.Error)
	UpdateSessionActivity(ctx context.Context, sessionId string, lastActivityAt time.Time) domain.Error
	UpdateSessionTokens(ctx context.Context, sessionId string, accessTokenEnc, refreshTokenEnc []byte, accessTokenExpiresAt time.Time) domain.Error
	InvalidateSession(ctx context.Context, sessionId string, reason string) domain.Error
	InvalidateAllUserSessions(ctx context.Context, userId string, reason string) domain.Error
	DeleteExpiredSessions(ctx context.Context) (int64, domain.Error)
	FindSessionsNeedingTokenRefresh(ctx context.Context, beforeTime time.Time) ([]domain.Session, domain.Error)
}
