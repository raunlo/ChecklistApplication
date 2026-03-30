//go:generate go run github.com/google/wire/cmd/wire@latest
//go:build wireinject
// +build wireinject

package deployment

import (
	guardrail "com.raunlo.checklist/internal/core/guard_rail"
	"com.raunlo.checklist/internal/core/notification"
	coreRepo "com.raunlo.checklist/internal/core/repository"
	"com.raunlo.checklist/internal/core/service"
	"com.raunlo.checklist/internal/job"
	"com.raunlo.checklist/internal/repository"
	"com.raunlo.checklist/internal/repository/connection"
	"com.raunlo.checklist/internal/server"
	"com.raunlo.checklist/internal/server/auth"
	authV1 "com.raunlo.checklist/internal/server/v1/auth"
	checklistV1 "com.raunlo.checklist/internal/server/v1/checklist"
	checklistItemV1 "com.raunlo.checklist/internal/server/v1/checklistItem"
	"com.raunlo.checklist/internal/server/v1/sse"
	templateV1 "com.raunlo.checklist/internal/server/v1/template"
	userV1 "com.raunlo.checklist/internal/server/v1/user"
	wire "github.com/google/wire"
)

// provideBaseUrl extracts the baseUrl from ServerConfiguration
func provideBaseUrl(config ServerConfiguration) auth.BaseUrl {
	return auth.BaseUrl(config.BaseUrl)
}

// provideFrontendUrl extracts the frontendUrl from ServerConfiguration
func provideFrontendUrl(config ServerConfiguration) auth.FrontendUrl {
	return auth.FrontendUrl(config.FrontendUrl)
}

// provideTokenEncryptor creates a TokenEncryptor from the SessionAuthConfiguration
// Panics on error since encryption key is required for the app to function
func provideTokenEncryptor(config SessionAuthConfiguration) auth.TokenEncryptor {
	encryptor, err := auth.NewTokenEncryptor(config.EncryptionKey)
	if err != nil {
		panic("Failed to create token encryptor: " + err.Error())
	}
	return encryptor
}

// provideGoogleOAuthConfig creates a GoogleOAuthConfig from the configuration
func provideGoogleOAuthConfig(googleConfig GoogleSSOConfiguration, serverConfig ServerConfiguration) *auth.GoogleOAuthConfig {
	return &auth.GoogleOAuthConfig{
		ClientID:     googleConfig.ClientID,
		ClientSecret: googleConfig.ClientSecret,
		RedirectURL:  serverConfig.BaseUrl + "/api/v1/auth/google/callback",
	}
}

// provideCleanupJob creates the cleanup job with default configuration
func provideCleanupJob(repo coreRepo.IChecklistItemsRepository) *job.CleanupJob {
	return job.NewCleanupJob(repo, job.DefaultCleanupJobConfig())
}

func Init(configuration ApplicationConfiguration) Application {
	panic(wire.Build(
		GetGinRouter,
		CreateApplication,
		server.NewRoutes,
		provideBaseUrl,
		provideFrontendUrl,
		provideCleanupJob,
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
			service.CreateRebalanceService,
			repository.CreateChecklistItemRepository,
			notification.NewNotificationService,
			notification.NewBroker,
		),
		// user resource set (GDPR endpoints)
		wire.NewSet(
			userV1.NewUserController,
			service.NewUserService,
			repository.NewUserRepository,
		),
		// auth resource set (session-based authentication)
		wire.NewSet(
			authV1.NewAuthController,
			service.NewAuthSessionService,
			wire.Bind(new(auth.SessionValidator), new(service.IAuthSessionService)),
			repository.NewSessionRepository,
			provideTokenEncryptor,
			provideGoogleOAuthConfig,
		),
		// template resource set
		wire.NewSet(
			templateV1.NewTemplateController,
			templateV1.NewTemplateDtoMapper,
			service.CreateTemplateService,
			service.CreateTemplateInviteService,
			repository.CreateTemplateRepository,
			repository.CreateTemplateInviteRepository,
			guardrail.NewTemplateOwnershipCheckerService,
		),
		connection.NewDatabaseConnection,
		sse.NewSSEController,
		wire.FieldsOf(new(ApplicationConfiguration), "DatabaseConfiguration"),
		wire.FieldsOf(new(ApplicationConfiguration), "ServerConfiguration"),
		wire.FieldsOf(new(ApplicationConfiguration), "CorsConfiguration"),
		wire.FieldsOf(new(ApplicationConfiguration), "GoogleSSOConfiguration"),
		wire.FieldsOf(new(ApplicationConfiguration), "SessionAuthConfiguration"),
	))
}
