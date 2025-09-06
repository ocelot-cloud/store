package apps

import (
	"fmt"
	"net/http"
	"ocelot/store/tools"
	"ocelot/store/users"
	"strconv"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	"github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
)

func AppCreationHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	appString, err := validation.ReadBody[store.AppNameString](w, r)
	if err != nil {
		return
	}

	if !users.UserRepo.DoesUserExist(user) {
		Logger.Info("user tried to create app but it does not exist", tools.UserField, user, tools.AppField, appString)
		http.Error(w, "user does not exists", http.StatusBadRequest)
		return
	}

	if appString.Value == "ocelotcloud" {
		Logger.Info("user tried to create app but it is reserved", tools.UserField, user, tools.AppField, appString)
		http.Error(w, "app name is reserved", http.StatusBadRequest)
		return
	}

	_, err = AppRepo.GetAppId(user, appString.Value)
	if err == nil {
		Logger.Info("user tried to create app but it already exists", tools.UserField, user, tools.AppField, appString)
		http.Error(w, "app already exists", http.StatusBadRequest)
		return
	}

	err = AppRepo.CreateApp(user, appString.Value)
	if err != nil {
		Logger.Error("user tried to create app but it failed", tools.UserField, user, tools.AppField, appString, deepstack.ErrorField, err)
		http.Error(w, "app creation failed", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	Logger.Info("user created app", tools.UserField, user, tools.AppField, appString)
}

func AppDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	appId, err := ReadBodyAsStringNumber(w, r)
	if err != nil {
		return
	}

	if !AppRepo.IsAppOwner(user, appId) {
		Logger.Warn("user tried to delete app with ID but does not own it", tools.UserField, user, tools.AppIdField, appId)
		http.Error(w, "you do not own this app", http.StatusBadRequest)
		return
	}

	err = AppRepo.DeleteApp(appId)
	if err != nil {
		Logger.Error("user tried to delete app with ID but it failed", tools.UserField, user, tools.AppIdField, appId)
		http.Error(w, "app deletion failed", http.StatusBadRequest)
		return
	}

	Logger.Info("user deleted app with ID", tools.UserField, user, tools.AppIdField, appId)
	w.WriteHeader(http.StatusOK)
}

func ReadBodyAsStringNumber(w http.ResponseWriter, r *http.Request) (int, error) {
	appIdString, err := validation.ReadBody[store.NumberString](w, r)
	if err != nil {
		return -1, fmt.Errorf("")
	}
	appId, err := strconv.Atoi(appIdString.Value)
	if err != nil {
		Logger.Warn("request body string conversion error", tools.AppIdField, appIdString)
		http.Error(w, "invalid input", http.StatusBadRequest)
		return -1, fmt.Errorf("")
	}
	return appId, nil
}

func AppGetListHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)

	list, err := AppRepo.GetAppList(user)
	if err != nil {
		Logger.Warn("error getting app list", deepstack.ErrorField, err)
		http.Error(w, "error getting app list", http.StatusBadRequest)
	}

	Logger.Info("got apps of user", tools.UserField, user)
	utils.SendJsonResponse(w, list)
}

func SearchForAppsHandler(w http.ResponseWriter, r *http.Request) {
	appSearchRequest, err := validation.ReadBody[store.AppSearchRequest](w, r)
	if err != nil {
		return
	}

	apps, err := AppRepo.SearchForApps(*appSearchRequest)
	if err != nil {
		Logger.Warn("error finding apps", deepstack.ErrorField, err)
		http.Error(w, "error finding apps", http.StatusBadRequest)
		return
	}

	utils.SendJsonResponse(w, apps)
}
