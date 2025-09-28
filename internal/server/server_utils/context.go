package serverutils

import (
	"context"

	"com.raunlo.checklist/internal/core/domain"
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
		return ctx
	}
}
