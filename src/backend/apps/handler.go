package apps

import (
	"fmt"
	"net/http"
	"ocelot/store/tools"
	"ocelot/store/users"
	"strconv"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
)

type AppsHandler struct {
	AppRepo    AppRepository
	UserRepo   users.UserRepository
	AppService *AppServiceImpl
}

func (a *AppsHandler) AppCreationHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	appString, err := validation.ReadBody[store.AppNameString](w, r)
	if err != nil {
		return
	}
	err = a.AppService.CreateAppWithChecks(user.Id, appString.Value)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(AppNameReservedError, AppAlreadyExistsError), err)
	}
}

func (a *AppsHandler) AppDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	appId, err := ReadBodyAsStringNumber(w, r)
	if err != nil {
		return
	}
	err = a.AppService.DeleteAppWithChecks(user.Id, appId)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(YouDoNotOwnThisAppError), err)
	}
}

func ReadBodyAsStringNumber(w http.ResponseWriter, r *http.Request) (int, error) {
	appIdString, err := validation.ReadBody[store.NumberString](w, r)
	if err != nil {
		return -1, fmt.Errorf("")
	}
	appId, err := strconv.Atoi(appIdString.Value)
	if err != nil {
		u.Logger.Warn("request body string conversion error", tools.AppIdField, appIdString)
		http.Error(w, "invalid input", http.StatusBadRequest)
		return -1, fmt.Errorf("")
	}
	return appId, nil
}

func (a *AppsHandler) AppGetListHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)

	list, err := a.AppRepo.GetAppList(user.Id)
	if err != nil {
		u.Logger.Warn("error getting app list", deepstack.ErrorField, err)
		http.Error(w, "error getting app list", http.StatusBadRequest)
	}

	u.Logger.Info("got apps of user", tools.UserField, user)
	u.SendJsonResponse(w, list)
}

func (a *AppsHandler) SearchForAppsHandler(w http.ResponseWriter, r *http.Request) {
	appSearchRequest, err := validation.ReadBody[store.AppSearchRequest](w, r)
	if err != nil {
		return
	}

	apps, err := a.AppRepo.SearchForApps(*appSearchRequest)
	if err != nil {
		u.Logger.Warn("error finding apps", deepstack.ErrorField, err)
		http.Error(w, "error finding apps", http.StatusBadRequest)
		return
	}

	u.SendJsonResponse(w, apps)
}
