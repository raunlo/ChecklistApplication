package dbo

import (
	"time"

	"com.raunlo.checklist/internal/core/domain"
)

type WorkspaceDBO struct {
	Id          uint64    `primaryKey:"id"`
	OwnerUserId string    `db:"owner_user_id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	MemberCount int       `db:"member_count"`
	IsOwner     bool      `db:"is_owner"`
	IsDefault   bool      `db:"is_default"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (d *WorkspaceDBO) ToDomain() domain.Workspace {
	return domain.Workspace{
		Id:          uint(d.Id),
		OwnerUserId: d.OwnerUserId,
		Name:        d.Name,
		Description: d.Description,
		MemberCount: d.MemberCount,
		IsOwner:     d.IsOwner,
		IsDefault:   d.IsDefault,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func (d *WorkspaceDBO) FromDomain(w domain.Workspace) {
	d.Id = uint64(w.Id)
	d.OwnerUserId = w.OwnerUserId
	d.Name = w.Name
	d.Description = w.Description
	d.IsDefault = w.IsDefault
	d.CreatedAt = w.CreatedAt
	d.UpdatedAt = w.UpdatedAt
}

type WorkspaceMemberDBO struct {
	Id      uint64  `primaryKey:"id"`
	UserId  string  `db:"user_id"`
	Name    *string `db:"name"`
	IsOwner bool    `db:"is_owner"`
}

func (d WorkspaceMemberDBO) ToDomain(workspaceId uint) domain.WorkspaceMember {
	return domain.WorkspaceMember{
		MemberId:    uint(d.Id),
		WorkspaceId: workspaceId,
		UserId:      d.UserId,
		Name:        d.Name,
		IsOwner:     d.IsOwner,
	}
}

type WorkspaceInviteDBO struct {
	Id          uint64     `primaryKey:"id"`
	WorkspaceId uint64     `db:"workspace_id"`
	Name        *string    `db:"name"`
	InviteToken string     `db:"invite_token"`
	CreatedBy   string     `db:"created_by"`
	CreatedAt   time.Time  `db:"created_at"`
	ExpiresAt   *time.Time `db:"expires_at"`
	ClaimedBy   *string    `db:"claimed_by"`
	ClaimedAt   *time.Time `db:"claimed_at"`
	IsSingleUse bool       `db:"is_single_use"`
}

func MapWorkspaceInviteDboToDomain(d WorkspaceInviteDBO) domain.WorkspaceInvite {
	return domain.WorkspaceInvite{
		Id:          uint(d.Id),
		WorkspaceId: uint(d.WorkspaceId),
		Name:        d.Name,
		InviteToken: d.InviteToken,
		CreatedBy:   d.CreatedBy,
		CreatedAt:   d.CreatedAt,
		ExpiresAt:   d.ExpiresAt,
		ClaimedBy:   d.ClaimedBy,
		ClaimedAt:   d.ClaimedAt,
		IsSingleUse: d.IsSingleUse,
	}
}
