package error

import (
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
)

const (
	errorMessageChecklistNotFound      = "Checklist(id=%d) not found"
	errorMessageInviteNotFound          = "Invite not found"
	errorMessageInviteExpired           = "This invite link has expired"
	errorMessageInviteAlreadyClaimed    = "This invite has already been used"
	errorMessageNotChecklistOwner       = "You must be the checklist owner to perform this action"
	errorMessageInvalidInviteToken      = "Invalid invite token"
)

func NewChecklistNotFoundError(checklistId uint) domain.Error {
	return domain.NewError(fmt.Sprintf(errorMessageChecklistNotFound, checklistId), 404)
}

func NewInviteNotFoundError() domain.Error {
	return domain.NewError(errorMessageInviteNotFound, 404)
}

func NewInviteExpiredError() domain.Error {
	return domain.NewError(errorMessageInviteExpired, 400)
}

func NewInviteAlreadyClaimedError() domain.Error {
	return domain.NewError(errorMessageInviteAlreadyClaimed, 400)
}

func NewNotChecklistOwnerError(checklistId uint) domain.Error {
	return domain.NewError(fmt.Sprintf("You must be the owner of checklist %d to perform this action", checklistId), 403)
}

func NewInvalidInviteTokenError() domain.Error {
	return domain.NewError(errorMessageInvalidInviteToken, 400)
}
