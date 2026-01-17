package user

import (
	"context"
	"log"
	"time"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/core/service"
)

type IUserController interface {
	StrictServerInterface
}

type userControllerImpl struct {
	userService service.IUserService
}

func NewUserController(userService service.IUserService) IUserController {
	return &userControllerImpl{
		userService: userService,
	}
}

// DeleteAccount implements GDPR Right to Erasure (Article 17)
func (ctrl *userControllerImpl) DeleteAccount(ctx context.Context, request DeleteAccountRequestObject) (DeleteAccountResponseObject, error) {
	userId, _ := domain.GetUserIdFromContext(ctx)

	// Require explicit confirmation
	if !request.Params.Confirm {
		return DeleteAccount400JSONResponse{
			Message: "Add ?confirm=true to delete account",
		}, nil
	}

	log.Printf("User(id=%s) requested account deletion", domain.GetHashedUserIdFromContext(ctx))

	// Delete account (cascades to all checklists, items, rows, shares, invites)
	if err := ctrl.userService.DeleteAccount(ctx, userId); err != nil {
		log.Printf("Failed to delete account for user(id=%s): %v", domain.GetHashedUserIdFromContext(ctx), err)
		return DeleteAccount500JSONResponse{
			Message: "Failed to delete account",
		}, nil
	}

	log.Printf("Successfully deleted account for user(id=%s)", domain.GetHashedUserIdFromContext(ctx))

	successMsg := "Account deleted successfully"
	return DeleteAccount200JSONResponse{
		Message: &successMsg,
	}, nil
}

// ExportUserData implements GDPR Data Portability (Article 20)
func (ctrl *userControllerImpl) ExportUserData(ctx context.Context, request ExportUserDataRequestObject) (ExportUserDataResponseObject, error) {
	userId, _ := domain.GetUserIdFromContext(ctx)

	log.Printf("User(id=%s) requested data export", domain.GetHashedUserIdFromContext(ctx))

	export, err := ctrl.userService.ExportUserData(ctx, userId)
	if err != nil {
		log.Printf("Failed to export data for user(id=%s): %v", domain.GetHashedUserIdFromContext(ctx), err)
		return ExportUserData500JSONResponse{
			Message: "Failed to export user data",
		}, nil
	}

	log.Printf("Successfully exported data for user(id=%s): %d checklists", domain.GetHashedUserIdFromContext(ctx), len(export.Checklists))

	// Convert domain model to API response
	response := convertToUserDataExport(export)

	return ExportUserData200JSONResponse(response), nil
}

func convertToUserDataExport(export *domain.UserDataExport) UserDataExport {
	result := UserDataExport{
		UserId:     export.UserId,
		ExportedAt: export.ExportedAt,
		Checklists: make([]struct {
			CreatedAt *time.Time `json:"createdAt,omitempty"`
			Id        *uint      `json:"id,omitempty"`
			Items     *[]struct {
				Completed   *bool   `json:"completed,omitempty"`
				Id          *uint   `json:"id,omitempty"`
				Name        *string `json:"name,omitempty"`
				OrderNumber *int    `json:"orderNumber,omitempty"`
				Rows        *[]struct {
					Completed *bool   `json:"completed,omitempty"`
					Id        *uint   `json:"id,omitempty"`
					Name      *string `json:"name,omitempty"`
				} `json:"rows,omitempty"`
			} `json:"items,omitempty"`
			Name   *string `json:"name,omitempty"`
			Shares *[]struct {
				PermissionLevel  *string    `json:"permissionLevel,omitempty"`
				SharedAt         *time.Time `json:"sharedAt,omitempty"`
				SharedWithUserId *string    `json:"sharedWithUserId,omitempty"`
			} `json:"shares,omitempty"`
		}, len(export.Checklists)),
	}

	for i, checklist := range export.Checklists {
		result.Checklists[i].Id = &checklist.Id
		result.Checklists[i].Name = &checklist.Name
		result.Checklists[i].CreatedAt = &checklist.CreatedAt

		// Convert items
		if len(checklist.Items) > 0 {
			items := make([]struct {
				Completed   *bool   `json:"completed,omitempty"`
				Id          *uint   `json:"id,omitempty"`
				Name        *string `json:"name,omitempty"`
				OrderNumber *int    `json:"orderNumber,omitempty"`
				Rows        *[]struct {
					Completed *bool   `json:"completed,omitempty"`
					Id        *uint   `json:"id,omitempty"`
					Name      *string `json:"name,omitempty"`
				} `json:"rows,omitempty"`
			}, len(checklist.Items))

			for j, item := range checklist.Items {
				items[j].Id = &item.Id
				items[j].Name = &item.Name
				items[j].Completed = &item.Completed
				items[j].OrderNumber = &item.OrderNumber

				// Convert rows
				if len(item.Rows) > 0 {
					rows := make([]struct {
						Completed *bool   `json:"completed,omitempty"`
						Id        *uint   `json:"id,omitempty"`
						Name      *string `json:"name,omitempty"`
					}, len(item.Rows))

					for k, row := range item.Rows {
						rows[k].Id = &row.Id
						rows[k].Name = &row.Name
						rows[k].Completed = &row.Completed
					}
					items[j].Rows = &rows
				}
			}
			result.Checklists[i].Items = &items
		}

		// Convert shares
		if len(checklist.Shares) > 0 {
			shares := make([]struct {
				PermissionLevel  *string    `json:"permissionLevel,omitempty"`
				SharedAt         *time.Time `json:"sharedAt,omitempty"`
				SharedWithUserId *string    `json:"sharedWithUserId,omitempty"`
			}, len(checklist.Shares))

			for j, share := range checklist.Shares {
				shares[j].SharedWithUserId = &share.SharedWithUserId
				shares[j].PermissionLevel = &share.PermissionLevel
				shares[j].SharedAt = &share.SharedAt
			}
			result.Checklists[i].Shares = &shares
		}
	}

	return result
}
