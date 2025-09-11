package versions

import (
	"encoding/json"
	"net/http"
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"
	"strconv"
	"strings"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
)

type VersionsHandler struct {
	VersionRepo    VersionRepository
	AppRepo        apps.AppRepository
	UserRepo       users.UserRepository
	VersionService *VersionService
}

// TODO !! too long, shift to service, maybe simplify?
func (v *VersionsHandler) VersionUploadHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	r.Body = http.MaxBytesReader(w, r.Body, tools.MaxPayloadSize)
	defer u.Close(r.Body)

	var versionUpload store.VersionUpload
	err := json.NewDecoder(r.Body).Decode(&versionUpload)
	if err != nil {
		if err.Error() == "http: request body too large" {
			u.Logger.Info("version upload version content of user was too large", tools.UserField, user)
			http.Error(w, "version content too large, the limit is 1MB", http.StatusBadRequest)
			return
		} else {
			u.Logger.Info("version upload request body of user was invalid", tools.UserField, user, deepstack.ErrorField, err)
			http.Error(w, "could not decode request body", http.StatusBadRequest)
			return
		}
	}

	err = validation.ValidateStruct(versionUpload)
	if err != nil {
		u.Logger.Info("version upload of user failed", tools.UserField, user, deepstack.ErrorField, err)
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	err = v.UserRepo.IsThereEnoughSpaceToAddVersion(user, len(versionUpload.Content))
	if err != nil {
		if strings.HasPrefix(err.Error(), users.NotEnoughSpacePrefix) {
			u.Logger.Info("version upload of user failed: not enough space", tools.UserField, user)
			http.Error(w, err.Error(), http.StatusInsufficientStorage)
			return
		} else {
			http.Error(w, "internal error", http.StatusBadRequest)
			return
		}
	}

	appId, err := strconv.Atoi(versionUpload.AppId)
	if err != nil {
		u.Logger.Info("user tried to upload version to app, but app ID is not a number", tools.UserField, user, tools.VersionField, versionUpload.Version, tools.AppIdField, versionUpload.AppId)
		http.Error(w, "could not convert to number", http.StatusBadRequest)
		return
	}

	if !v.AppRepo.DoesAppExist(appId) {
		u.Logger.Info("user tried to upload version to app, but app does not exist", tools.UserField, user, tools.VersionField, versionUpload.Version, tools.AppIdField, versionUpload.AppId)
		http.Error(w, "app does not exist", http.StatusBadRequest)
		return
	}

	if !v.AppRepo.DoesUserOwnApp(user, appId) {
		u.Logger.Warn("user tried to delete app but does not own it", tools.UserField, user, tools.AppIdField, versionUpload.AppId)
		http.Error(w, "you do not own this app", http.StatusBadRequest)
		return
	}

	appName, err := v.AppRepo.GetAppName(appId)
	if err != nil {
		u.Logger.Error("getting app name failed", deepstack.ErrorField, err)
		http.Error(w, "internal error", http.StatusBadRequest)
		return
	}

	maintainerName, err := v.AppRepo.GetMaintainerName(appId)
	if err != nil {
		u.Logger.Error("getting maintainer name failed", deepstack.ErrorField, err)
		http.Error(w, "internal error", http.StatusBadRequest)
		return
	}

	// TODO !! add deepstack errors
	err = validation.ValidateVersion(versionUpload.Content, maintainerName, appName)
	if err != nil {
		// TODO !! expected error: "zip: not a valid zip file" -> make this a an error in "shared" for reuse?
		u.WriteResponseError(w, u.MapOf("zip: not a valid zip file"), err, tools.UserField, user)
		return
	}

	_, err = v.VersionRepo.GetVersionId(appId, versionUpload.Version)
	if err == nil {
		u.Logger.Info("user tried to upload version to app, but version already exists", tools.UserField, user, tools.VersionField, versionUpload.Version, tools.AppIdField, versionUpload.AppId)
		http.Error(w, "version already exists", http.StatusBadRequest)
		return
	}

	err = v.VersionRepo.CreateVersion(appId, versionUpload.Version, versionUpload.Content)
	if err != nil {
		u.Logger.Error("creating version failed", deepstack.ErrorField, err)
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	u.Logger.Info("version was uploaded to app by user", tools.VersionField, versionUpload.Version, tools.AppIdField, versionUpload.AppId, tools.UserField, user)
	w.WriteHeader(http.StatusOK)
}

func (v *VersionsHandler) VersionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r) // TODO !! should be a user, not only a string
	versionId, err := apps.ReadBodyAsStringNumber(w, r)
	if err != nil {
		return
	}
	err = v.VersionService.DeleteVersionWithChecks(user, versionId)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(NotOwningThisVersionError, VersionDoesNotExistError), err, tools.UserField, user, tools.VersionIdField, versionId)
	}
}

func (v *VersionsHandler) GetVersionsHandler(w http.ResponseWriter, r *http.Request) {
	appId, err := apps.ReadBodyAsStringNumber(w, r)
	if err != nil {
		return
	}

	if !v.AppRepo.DoesAppExist(appId) {
		u.Logger.Info("someone tried to list versions but app does not exist", tools.AppIdField, appId)
		http.Error(w, "app does not exist", http.StatusBadRequest)
		return
	}

	versionsList, err := v.VersionRepo.ListVersionsOfApp(appId)
	if err != nil {
		u.Logger.Error("getting version list failed for app", tools.AppIdField, appId)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u.SendJsonResponse(w, versionsList)
}

func (v *VersionsHandler) VersionDownloadHandler(w http.ResponseWriter, r *http.Request) {
	versionId, err := apps.ReadBodyAsStringNumber(w, r)
	if err != nil {
		return
	}

	// TODO !! shift check to service
	doesExist, err := v.VersionRepo.DoesVersionExist(versionId)
	if err != nil {
		u.Logger.Error("error when checking if version exists", deepstack.ErrorField, err)
		http.Error(w, "error when checking if version exists", http.StatusBadRequest)
		return
	}
	if !doesExist {
		u.Logger.Info("version does not exist", tools.VersionIdField, versionId)
		http.Error(w, "version does not exist", http.StatusBadRequest)
		return
	}

	versionInfo, err := v.VersionRepo.GetFullVersionInfo(versionId)
	if err != nil {
		u.Logger.Error("error when accessing version info", deepstack.ErrorField, err)
		http.Error(w, "error when accessing version info", http.StatusBadRequest)
		return
	}

	u.SendJsonResponse(w, versionInfo)
}
