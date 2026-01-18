package domain

import "time"

type Session struct {
	Id                    uint
	SessionId             string
	UserId                string
	AccessTokenEncrypted  []byte
	RefreshTokenEncrypted []byte
	AccessTokenExpiresAt  time.Time
	CreatedAt             time.Time
	LastActivityAt        time.Time
	ExpiresAt             time.Time
	IsInvalidated         bool
	InvalidatedAt         *time.Time
	InvalidationReason    *string
}

const (
	MaxSessionLifetime = 30 * 24 * time.Hour // 30 days
	IdleTimeout        = 7 * 24 * time.Hour  // 7 days
)
