package apps

import (
	"fmt"
	"github.com/ocelot-cloud/shared/store"
	"github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
	"net/http"
	"ocelot/store/tools"
	"ocelot/store/users"
	"strconv"
)

func AppCreationHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	appString, err := validation.ReadBody[store.AppNameString](w, r)
	if err != nil {
		return
	}

	if !users.UserRepo.DoesUserExist(user) {
		tools.Logger.InfoF("user '%s' tried to create app '%s' but it does not exist", user, appString)
		http.Error(w, "user does not exists", http.StatusNotFound)
		return
	}

	if appString.Value == "ocelotcloud" {
		tools.Logger.InfoF("user '%s' tried to create app '%s' but it is reserved", user, appString)
		http.Error(w, "app name is reserved", http.StatusBadRequest)
		return
	}

	_, err = AppRepo.GetAppId(user, appString.Value)
	if err == nil {
		tools.Logger.InfoF("user '%s' tried to create app '%s' but it already exists", user, appString)
		http.Error(w, "app already exists", http.StatusConflict)
		return
	}

	err = AppRepo.CreateApp(user, appString.Value)
	if err != nil {
		tools.Logger.ErrorF("user '%s' tried to create app '%s' but it failed: %v", user, appString, err)
		http.Error(w, "app creation failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	tools.Logger.InfoF("user '%s' created app '%s'", user, appString)
}

func AppDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	appId, err := ReadBodyAsStringNumber(w, r)
	if err != nil {
		return
	}

	if !AppRepo.IsAppOwner(user, appId) {
		tools.Logger.WarnF("user '%s' tried to delete app with ID '%d' but does not own it", user, appId)
		http.Error(w, "you do not own this app", http.StatusUnauthorized)
		return
	}

	err = AppRepo.DeleteApp(appId)
	if err != nil {
		tools.Logger.ErrorF("user '%s' tried to delete app with ID '%d' but it failed", user, appId)
		http.Error(w, "app deletion failed", http.StatusInternalServerError)
		return
	}

	tools.Logger.InfoF("user '%s' deleted app with ID '%d'", user, appId)
	w.WriteHeader(http.StatusOK)
}

func ReadBodyAsStringNumber(w http.ResponseWriter, r *http.Request) (int, error) {
	appIdString, err := validation.ReadBody[store.NumberString](w, r)
	if err != nil {
		return -1, fmt.Errorf("")
	}
	appId, err := strconv.Atoi(appIdString.Value)
	if err != nil {
		tools.Logger.WarnF("request body string conversion error: %v", appIdString)
		http.Error(w, "invalid input", http.StatusBadRequest)
		return -1, fmt.Errorf("")
	}
	return appId, nil
}

func AppGetListHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)

	list, err := AppRepo.GetAppList(user)
	if err != nil {
		tools.Logger.WarnF("error getting app list: %v", err)
		http.Error(w, "error getting app list", http.StatusInternalServerError)
	}

	tools.Logger.InfoF("got apps of user '%s'", user)
	utils.SendJsonResponse(w, list)
}

func SearchForAppsHandler(w http.ResponseWriter, r *http.Request) {
	appSearchRequest, err := validation.ReadBody[store.AppSearchRequest](w, r)
	if err != nil {
		return
	}

	apps, err := AppRepo.SearchForApps(*appSearchRequest)
	if err != nil {
		tools.Logger.WarnF("error finding apps: %v", err)
		http.Error(w, "error finding apps", http.StatusInternalServerError)
		return
	}

	utils.SendJsonResponse(w, apps)
}
