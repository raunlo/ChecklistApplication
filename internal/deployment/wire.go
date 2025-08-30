//go:generate go run github.com/google/wire/cmd/wire@latest
//go:build wireinject
// +build wireinject

package deployment

import (
	"com.raunlo.checklist/internal/core/service"
	"com.raunlo.checklist/internal/repository"
	"com.raunlo.checklist/internal/repository/connection"
	"com.raunlo.checklist/internal/server"
	checklistV1 "com.raunlo.checklist/internal/server/v1/checklist"
	checklistItemV1 "com.raunlo.checklist/internal/server/v1/checklistItem"
	wire "github.com/google/wire"
)

func Init(configuration ApplicationConfiguration) Application {
	wire.Build(
		GetGinRouter,
		CreateApplication,
		server.NewRoutes,
		// checklist resource set
		wire.NewSet(
			checklistV1.NewChecklistController,
			service.CreateChecklistService,
			repository.CreateChecklistRepository,
		),
		// checklist item resource set
		wire.NewSet(
			checklistItemV1.NewChecklistItemController,
			service.CreateChecklistItemService,
			repository.CreateChecklistItemRepository,
		),
		// checklist item template resource set
		//wire.NewSet(controllerMapper.NewChecklistItemTemplateDtoMapper,
		//	controllers.CreateChecklistItemTemplateController,
		//	service.CreateChecklistItemTemplateService,
		//	repository.CreateChecklistItemTemplateRepository),
		//controllers.CreateUpdateOrderController,
		connection.NewDatabaseConnection,
		wire.FieldsOf(new(ApplicationConfiguration), "DatabaseConfiguration"),
		wire.FieldsOf(new(ApplicationConfiguration), "ServerConfiguration"),
		wire.FieldsOf(new(ApplicationConfiguration), "CorsConfiguration"),
	)
	return Application{}
}
