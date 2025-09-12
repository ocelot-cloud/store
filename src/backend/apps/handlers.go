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
	AppRepo  AppRepository
	UserRepo users.UserRepository
}

func (a *AppsHandler) AppCreationHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	appString, err := validation.ReadBody[store.AppNameString](w, r)
	if err != nil {
		return
	}

	if !a.UserRepo.DoesUserExist(user.UserName) {
		u.Logger.Info("user tried to create app but it does not exist", tools.UserField, user, tools.AppField, appString)
		http.Error(w, "user does not exists", http.StatusBadRequest)
		return
	}

	if appString.Value == "ocelotcloud" {
		u.Logger.Info("user tried to create app but it is reserved", tools.UserField, user, tools.AppField, appString)
		http.Error(w, "app name is reserved", http.StatusBadRequest)
		return
	}

	_, err = a.AppRepo.GetAppId(user.UserName, appString.Value)
	if err == nil {
		u.Logger.Info("user tried to create app but it already exists", tools.UserField, user, tools.AppField, appString)
		http.Error(w, "app already exists", http.StatusBadRequest)
		return
	}

	err = a.AppRepo.CreateApp(user.UserName, appString.Value)
	if err != nil {
		u.Logger.Error("user tried to create app but it failed", tools.UserField, user, tools.AppField, appString, deepstack.ErrorField, err)
		http.Error(w, "app creation failed", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	u.Logger.Info("user created app", tools.UserField, user, tools.AppField, appString)
}

func (a *AppsHandler) AppDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	appId, err := ReadBodyAsStringNumber(w, r)
	if err != nil {
		return
	}

	if !a.AppRepo.DoesUserOwnApp(user.UserName, appId) {
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

	list, err := a.AppRepo.GetAppList(user.UserName)
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
