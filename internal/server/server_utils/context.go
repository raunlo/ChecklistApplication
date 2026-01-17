package serverutils

import (
	"context"
	"log"

	"com.raunlo.checklist/internal/core/domain"
	"com.raunlo.checklist/internal/server/auth"
	"github.com/gin-gonic/gin"
)

func CreateContext(ginContext context.Context) context.Context {
	castedGinContext, ok := ginContext.(*gin.Context)
	ctx := context.Background()
	if !ok {
		panic("invalid context type")
	}

	// Always add user ID to context (required for authentication)
	ctx = createContextWithUserId(castedGinContext, ctx)

	// Try to get the clientId from the header first
	clientId := castedGinContext.GetHeader("X-Client-Id")
	if clientId == "" {
		// For SSE connections, the clientId is passed as a query parameter
		clientId, _ = castedGinContext.GetQuery("clientId")
	}

	// Add client ID to context if present (optional for SSE filtering)
	if clientId != "" {
		ctx = context.WithValue(ctx, domain.ClientIdContextKey, clientId)
	}

	return ctx
}

func createContextWithUserId(ginContext *gin.Context, ctx context.Context) context.Context {
	userId, exists := auth.ExtractUserIdFromGinContext(ginContext)
	if exists {
		// Add both real and hashed user ID to context
		ctx = domain.AddUserIdToContext(ctx, userId)
	} else {
		log.Printf("Warning: User ID not found in context for request to %s", ginContext.Request.URL.Path)
	}
	return ctx
}
