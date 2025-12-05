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
	// Try to get the clientId from the header first
	clientId := castedGinContext.GetHeader("X-Client-Id")
	if clientId == "" {
		// For SSE connections, the clientId is passed as a query parameter
		clientId, _ = castedGinContext.GetQuery("clientId")
	}

	if clientId == "" {
		return ctx
	} else {
		ctx = context.WithValue(ctx, domain.ClientIdContextKey, clientId)
		ctx = createContextWithUserId(castedGinContext, ctx)
		return ctx
	}
}

func createContextWithUserId(ginContext *gin.Context, ctx context.Context) context.Context {
	userId, exists := auth.ExtractUserIdFromGinContext(ginContext)
	if exists {
		ctx = context.WithValue(ctx, domain.UserIdContextKey, userId)
	} else {
		log.Printf("Warning: User ID not found in context for request to %s", ginContext.Request.URL.Path)
	}
	return ctx
}
