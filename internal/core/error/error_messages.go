package error

import (
	"fmt"

	"com.raunlo.checklist/internal/core/domain"
)

const (
	errorMessageChecklistNotFound = "Checklist(id=%d) not found"
)

func NewChecklistNotFoundError(checklistId uint) domain.Error {
	return domain.NewError(fmt.Sprintf(errorMessageChecklistNotFound, checklistId), 404)
}
