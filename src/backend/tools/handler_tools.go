package tools

import (
	"net/http"
)

type ContextKey string

const UserCtxKey ContextKey = "user"

// GetUserFromContext Since only authenticated users are added to the context, it only works in protected handlers.
func GetUserFromContext(r *http.Request) string {
	return r.Context().Value(UserCtxKey).(string)
}
