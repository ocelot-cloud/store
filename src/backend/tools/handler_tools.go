package tools

import (
	"net/http"
)

type ContextKey string

const UserCtxKey ContextKey = "user"

// GetUserFromContext Since only authenticated users are added to the context, it only works in protected handlers.
func GetUserFromContext(r *http.Request) User {
	return r.Context().Value(UserCtxKey).(User)
}

type User struct {
	UserId            int
	UserName          string
	Email             string
	HashedPassword    string
	HashedCookieValue *string
	ExpirationDate    *string
	UsedSpace         int
}
