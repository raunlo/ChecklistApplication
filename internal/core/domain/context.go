package domain

var ClientIdContextKey = clientIdContextKey{}

type (
	userIdContextKey   struct{}
	clientIdContextKey = struct{}
)

var UserIdContextKey = userIdContextKey{}
