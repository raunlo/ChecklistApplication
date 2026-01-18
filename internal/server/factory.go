package server

import (
	"time"

	"com.raunlo.checklist/internal/server/auth"
	"com.raunlo.checklist/internal/server/middleware"
	v1 "com.raunlo.checklist/internal/server/v1"
	authV1 "com.raunlo.checklist/internal/server/v1/auth"
	checklistV1 "com.raunlo.checklist/internal/server/v1/checklist"
	checklistItemV1 "com.raunlo.checklist/internal/server/v1/checklistItem"
	"com.raunlo.checklist/internal/server/v1/legal"
	"com.raunlo.checklist/internal/server/v1/sse"
	userV1 "com.raunlo.checklist/internal/server/v1/user"
	"github.com/gin-gonic/gin"
)

type IRoutes interface {
	ConfigureRoutes()
}
type routes struct {
	engine                  *gin.Engine
	checklistController     checklistV1.IChecklistController
	checklistItemController checklistItemV1.IChecklistItemController
	sseController           sse.ISSEController
	userController          userV1.IUserController
	authController          *authV1.AuthController
	authSessionService      auth.SessionValidator
}

func NewRoutes(
	engine *gin.Engine,
	checklistController checklistV1.IChecklistController,
	checklistItemController checklistItemV1.IChecklistItemController,
	sseController sse.ISSEController,
	userController userV1.IUserController,
	authController *authV1.AuthController,
	authSessionService auth.SessionValidator,
) IRoutes {
	return &routes{
		engine:                  engine,
		checklistController:     checklistController,
		checklistItemController: checklistItemController,
		sseController:           sseController,
		userController:          userController,
		authController:          authController,
		authSessionService:      authSessionService,
	}
}

func (server *routes) ConfigureRoutes() {
	isProduction := gin.Mode() == gin.ReleaseMode

	// Public auth routes (no authentication required)
	authGroup := server.engine.Group("/api/v1/auth")
	authGroup.Use(auth.RateLimitMiddleware(10, time.Minute)) // Strict rate limit for auth endpoints
	authGroup.GET("/google/login", server.authController.InitiateGoogleLogin)
	authGroup.GET("/google/callback", server.authController.HandleGoogleCallback)

	// Dev login endpoint (only available in non-release mode)
	if !isProduction {
		authGroup.GET("/dev/login", server.authController.DevLogin)
	}

	// Public legal routes (no authentication required)
	publicGroup := server.engine.Group("/legal")
	publicGroup.GET("/privacy-policy", legal.GetPrivacyPolicy)
	publicGroup.GET("/privacy-policy-meta", legal.GetPrivacyPolicyJSON)

	// Protected routes (authentication required)
	protectedGroup := server.engine.Group("/")

	protectedGroup.Use(auth.RateLimitMiddleware(1000, time.Minute)) // 1000 requests per minute

	// Session-based authentication
	protectedGroup.Use(auth.SessionAuthMiddleware(server.authSessionService))

	// CSRF Protection (double-submit cookie pattern)
	if isProduction {
		protectedGroup.Use(middleware.SetCSRFTokenMiddleware(true, ""))
		protectedGroup.Use(middleware.CSRFMiddleware(true))
	} else {
		// Dev mode: use cross-subdomain cookie domain for local development
		// This allows app.dailychexly.local.com to read cookies set by api.dailychexly.local.com
		protectedGroup.Use(middleware.SetCSRFTokenMiddleware(false, ".dailychexly.local.com"))
		protectedGroup.Use(middleware.CSRFMiddleware(false))
	}

	// Protected auth endpoints (require authentication)
	authProtectedGroup := protectedGroup.Group("/api/v1/auth")
	authProtectedGroup.POST("/logout", server.authController.Logout)
	authProtectedGroup.GET("/session", server.authController.GetSession)

	// Register all V1 endpoints with authentication middleware
	v1.RegisterV1Endpoints(protectedGroup,
		v1.V1ProtectedEndpointsRegistrationRequest{
			ChecklistController:     server.checklistController,
			ChecklistItemController: server.checklistItemController,
			SSEController:           server.sseController,
			UserController:          server.userController,
		},
	)
}
