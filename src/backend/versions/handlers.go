package versions

import (
	"encoding/json"
	"github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
	"net/http"
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"
	"strconv"
	"strings"
)

func VersionUploadHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	r.Body = http.MaxBytesReader(w, r.Body, tools.MaxPayloadSize)
	defer utils.Close(r.Body)

	var versionUpload tools.VersionUpload
	err := json.NewDecoder(r.Body).Decode(&versionUpload)
	if err != nil {
		if err.Error() == "http: request body too large" {
			tools.Logger.Info("version upload version content of user '%s' was too large", user)
			http.Error(w, "version content too large, the limit is 1MB", http.StatusRequestEntityTooLarge)
			return
		} else {
			tools.Logger.Info("version upload request body of user '%s' was invalid: %v", user, err)
			http.Error(w, "could not decode request body", http.StatusBadRequest)
			return
		}
	}

	err = validation.ValidateStruct(versionUpload)
	if err != nil {
		tools.Logger.Info("version upload of user '%s' failed: %v", user, err)
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	err = users.UserRepo.IsThereEnoughSpaceToAddVersion(user, len(versionUpload.Content))
	if err != nil {
		if strings.HasPrefix(err.Error(), users.NotEnoughSpacePrefix) {
			tools.Logger.Info("version upload of user '%s' failed: not enough space", user)
			http.Error(w, err.Error(), http.StatusInsufficientStorage)
			return
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	appId, err := strconv.Atoi(versionUpload.AppId)
	if err != nil {
		tools.Logger.Info("user '%s' tried to upload version '%s' to app with ID '%s', but app ID is not a number", user, versionUpload.Version, versionUpload.AppId)
		http.Error(w, "could not convert to number", http.StatusBadRequest)
		return
	}

	if !apps.AppRepo.DoesAppExist(appId) {
		tools.Logger.Info("user '%s' tried to upload version '%s' to app with ID '%s', but app does not exist", user, versionUpload.Version, versionUpload.AppId)
		http.Error(w, "app does not exist", http.StatusNotFound)
		return
	}

	if !apps.AppRepo.IsAppOwner(user, appId) {
		tools.Logger.Warn("user '%s' tried to delete app with ID '%d' but does not own it", user, versionUpload.AppId)
		http.Error(w, "you do not own this app", http.StatusUnauthorized)
		return
	}

	appName, err := apps.AppRepo.GetAppName(appId)
	if err != nil {
		tools.Logger.Error("getting app name failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	maintainerName, err := apps.AppRepo.GetMaintainerName(appId)
	if err != nil {
		tools.Logger.Error("getting maintainer name failed: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	err = validation.ValidateVersion(versionUpload.Content, maintainerName, appName)
	if err != nil {
		tools.Logger.Info("version upload of user '%s' invalid: %v", user, err)
		http.Error(w, "invalid version: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err = VersionRepo.GetVersionId(appId, versionUpload.Version)
	if err == nil {
		tools.Logger.Info("user '%s' tried to upload version '%s' to app with ID '%s', but version already exists", user, versionUpload.Version, versionUpload.AppId)
		http.Error(w, "version already exists", http.StatusConflict)
		return
	}

	err = VersionRepo.CreateVersion(appId, versionUpload.Version, versionUpload.Content)
	if err != nil {
		tools.Logger.Error("creating version failed: %v", err)
		http.Error(w, "invalid input", http.StatusInternalServerError)
		return
	}

	tools.Logger.Info("version '%s' was uploaded to app with ID '%s' by user '%s'", versionUpload.Version, versionUpload.AppId, user)
	w.WriteHeader(http.StatusOK)
}

func VersionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	versionId, err := apps.ReadBodyAsStringNumber(w, r)
	if err != nil {
		return
	}

	if !VersionRepo.DoesVersionExist(versionId) {
		tools.Logger.Info("someone tried to delete version with ID '%d' but it does not exist", versionId)
		http.Error(w, "version does not exist", http.StatusNotFound)
		return
	}

	if !VersionRepo.IsVersionOwner(user, versionId) {
		tools.Logger.Warn("user '%s' tried to delete version with ID '%d' but does not own it", user, versionId)
		http.Error(w, "you do not own this version", http.StatusUnauthorized)
		return
	}

	err = VersionRepo.DeleteVersion(versionId)
	if err != nil {
		tools.Logger.Info("deleting version with ID '%d' failed: %v", versionId, err)
		http.Error(w, "invalid input", http.StatusInternalServerError)
		return
	}
	tools.Logger.Info("version with ID '%d' was deleted", versionId)
	http.Error(w, "version deleted", http.StatusOK)
}

func GetVersionsHandler(w http.ResponseWriter, r *http.Request) {
	appId, err := apps.ReadBodyAsStringNumber(w, r)
	if err != nil {
		return
	}

	if !apps.AppRepo.DoesAppExist(appId) {
		tools.Logger.Info("someone tried to list versions but app with ID '%d' does not exist", appId)
		http.Error(w, "app does not exist", http.StatusNotFound)
		return
	}

	versionsList, err := VersionRepo.GetVersionList(appId)
	if err != nil {
		tools.Logger.Error("getting version list failed for app with ID '%d'", appId)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJsonResponse(w, versionsList)
}

func VersionDownloadHandler(w http.ResponseWriter, r *http.Request) {
	versionId, err := apps.ReadBodyAsStringNumber(w, r)
	if err != nil {
		return
	}

	if !VersionRepo.DoesVersionExist(versionId) {
		tools.Logger.Info("version with ID '%d' does not exist", versionId)
		http.Error(w, "version does not exist", http.StatusNotFound)
		return
	}

	versionInfo, err := VersionRepo.GetFullVersionInfo(versionId)
	if err != nil {
		tools.Logger.Error("error when accessing version info: %v", err)
		http.Error(w, "error when accessing version info", http.StatusInternalServerError)
		return
	}

	utils.SendJsonResponse(w, versionInfo)
}
