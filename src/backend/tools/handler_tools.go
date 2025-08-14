package tools

import (
	"net/http"

	"github.com/ocelot-cloud/deepstack"
)

type ContextKey string

const UserCtxKey ContextKey = "user"

func HandleInvalidInput(w http.ResponseWriter, err error) {
	Logger.Info("invalid input", deepstack.ErrorField, err)
	http.Error(w, "invalid input", http.StatusBadRequest)
}

// GetUserFromContext Since only authenticated users are added to the context, it only works in protected handlers.
func GetUserFromContext(r *http.Request) string {
	return r.Context().Value(UserCtxKey).(string)
}
