package domain

import (
	"context"
)

var ClientIdContextKey = clientIdContextKey{}

type (
	userIdContextKey   struct{}
	clientIdContextKey = struct{}
)

var UserIdContextKey = userIdContextKey{}

func GetUserIdFromContext(ctx context.Context) (string, Error) {
	userIdValue := ctx.Value(UserIdContextKey)
	userId, ok := userIdValue.(string)
	if !ok || userId == "" {
		return "", NewError("User ID not found in context or invalid", 401)
	}
	return userId, nil
}
