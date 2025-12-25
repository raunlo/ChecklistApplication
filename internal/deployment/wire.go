//go:generate go run github.com/google/wire/cmd/wire@latest
//go:build wireinject
// +build wireinject

package deployment

import (
	guardrail "com.raunlo.checklist/internal/core/guard_rail"
	"com.raunlo.checklist/internal/core/notification"
	"com.raunlo.checklist/internal/core/service"
	"com.raunlo.checklist/internal/repository"
	"com.raunlo.checklist/internal/repository/connection"
	"com.raunlo.checklist/internal/server"
	"com.raunlo.checklist/internal/server/auth"
	checklistV1 "com.raunlo.checklist/internal/server/v1/checklist"
	checklistItemV1 "com.raunlo.checklist/internal/server/v1/checklistItem"
	"com.raunlo.checklist/internal/server/v1/sse"
	wire "github.com/google/wire"
)

// provideIDTokenValidator creates an IDTokenValidator from the Google SSO configuration
func provideIDTokenValidator(config GoogleSSOConfiguration) auth.IdtokenValidator {
	return auth.NewIDTokenValidator(config.ClientID)
}

// provideBaseUrl extracts the baseUrl from ServerConfiguration
func provideBaseUrl(config ServerConfiguration) string {
	return config.BaseUrl
}

func Init(configuration ApplicationConfiguration) Application {
	wire.Build(
		GetGinRouter,
		CreateApplication,
		server.NewRoutes,
		provideIDTokenValidator,
		provideBaseUrl,
		guardrail.NewChecklistOwnershipCheckerService,
		// checklist resource set
		wire.NewSet(
			checklistV1.NewChecklistController,
			service.CreateChecklistService,
			service.CreateChecklistInviteService,
			repository.CreateChecklistRepository,
			repository.CreateChecklistInviteRepository,
		),
		// checklist item resource set
		wire.NewSet(
			checklistItemV1.NewChecklistItemController,
			service.CreateChecklistItemService,
			repository.CreateChecklistItemRepository,
			notification.NewNotificationService,
			notification.NewBroker,
		),
		// checklist item template resource set
		// wire.NewSet(controllerMapper.NewChecklistItemTemplateDtoMapper,
		//	controllers.CreateChecklistItemTemplateController,
		//	service.CreateChecklistItemTemplateService,
		//	repository.CreateChecklistItemTemplateRepository),
		// controllers.CreateUpdateOrderController,
		connection.NewDatabaseConnection,
		sse.NewSSEController,
		wire.FieldsOf(new(ApplicationConfiguration), "DatabaseConfiguration"),
		wire.FieldsOf(new(ApplicationConfiguration), "ServerConfiguration"),
		wire.FieldsOf(new(ApplicationConfiguration), "CorsConfiguration"),
		wire.FieldsOf(new(ApplicationConfiguration), "GoogleSSOConfiguration"),
	)
	return Application{}
}
