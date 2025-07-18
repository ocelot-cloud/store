package tools

import (
	"github.com/ocelot-cloud/shared/utils"
	"net/http"
)

type ContextKey string

const UserCtxKey ContextKey = "user"

func HandleInvalidInput(w http.ResponseWriter, err error) {
	Logger.Info("invalid input", utils.ErrorField, err)
	http.Error(w, "invalid input", http.StatusBadRequest)
}

// GetUserFromContext Since only authenticated users are added to the context, it only works in protected handlers.
func GetUserFromContext(r *http.Request) string {
	return r.Context().Value(UserCtxKey).(string)
}
