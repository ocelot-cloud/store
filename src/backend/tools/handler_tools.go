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

// TODO !! I think we can add `db:` fields here, simplyfing sql queries. If so, also apply that to cloud
type User struct {
	Id                int
	Name              string
	Email             string
	HashedPassword    string
	HashedCookieValue *string
	ExpirationDate    *string
	UsedSpaceInBytes  int
}
