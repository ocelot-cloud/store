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

	isOwner, err := a.AppService.DoesUserOwnApp(user.Id, appId)
	if err != nil {
		u.Logger.Error("error when checking if user owns app", deepstack.ErrorField, err, tools.UserField, user, tools.AppIdField, appId)
		http.Error(w, "error when checking app ownership", http.StatusBadRequest)
		return
	}
	if !isOwner {
		u.Logger.Warn("user tried to delete app with ID but does not own it", tools.UserField, user, tools.AppIdField, appId)
		http.Error(w, "you do not own this app", http.StatusBadRequest)
		return
	}

	err = a.AppRepo.DeleteApp(appId)
	if err != nil {
		u.Logger.Error("user tried to delete app with ID but it failed", tools.UserField, user, tools.AppIdField, appId)
		http.Error(w, "app deletion failed", http.StatusBadRequest)
		return
	}

	u.Logger.Info("user deleted app with ID", tools.UserField, user, tools.AppIdField, appId)
	w.WriteHeader(http.StatusOK)
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
