package tools

import (
	"net/http"
)

// TODO !! to small, merge with others

type ContextKey string

const UserCtxKey ContextKey = "user"

// GetUserFromContext Since only authenticated users are added to the context, it only works in protected handlers.
func GetUserFromContext(r *http.Request) User {
	return r.Context().Value(UserCtxKey).(User)
}
