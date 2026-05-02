package domain

import "time"

type Workspace struct {
	Id          uint
	OwnerUserId string
	Name        string
	Description *string
	MemberCount int
	IsOwner     bool
	IsDefault   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkspaceMember struct {
	WorkspaceId uint
	UserId      string
	Email       string
	Name        *string
	IsOwner     bool
	JoinedAt    time.Time
}

type WorkspaceInvite struct {
	Id          uint
	WorkspaceId uint
	Name        *string
	InviteToken string
	CreatedBy   string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	ClaimedBy   *string
	ClaimedAt   *time.Time
	IsSingleUse bool
}
