package versions

import (
	"encoding/json"
	"net/http"
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
)

type VersionsHandler struct {
	VersionRepo    VersionRepository
	AppRepo        apps.AppRepository
	UserRepo       users.UserRepository
	VersionService *VersionService
	UserService    *users.UserServiceImpl
	AppService     *apps.AppServiceImpl
}

// TODO !! too long, shift to service, maybe simplify?
func (v *VersionsHandler) VersionUploadHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	r.Body = http.MaxBytesReader(w, r.Body, tools.MaxPayloadSize)
	defer u.Close(r.Body)
	var versionUpload store.VersionUploadDto
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
	err = v.VersionService.UploadVersion(user.Id, &versionUpload)
	if err != nil {
		// TODO !! should be put in "shared"
		// TODO !! space use case to be covered by component tests I guess? also NotOwningThisVersionError
		// TODO !! expected error: "zip: not a valid zip file" -> make this a an error in "shared" for reuse?

		expectedErros := u.MapOf("invalid input", users.NotEnoughSpacePrefix, NotOwningThisVersionError, "zip: not a valid zip file", VersionAlreadyExist, "app does not exist")
		u.WriteResponseError(w, expectedErros, err)
		return
	}
}

func (v *VersionsHandler) VersionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r) // TODO !! should be a user, not only a string
	versionId, err := apps.ReadBodyAsStringNumber(w, r)
	if err != nil {
		return
	}
	err = v.VersionService.DeleteVersionWithChecks(user.Id, versionId)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(NotOwningThisVersionError, VersionDoesNotExistError), err, tools.UserField, user, tools.VersionIdField, versionId)
	}
}

func (v *VersionsHandler) GetVersionsHandler(w http.ResponseWriter, r *http.Request) {
	appId, err := apps.ReadBodyAsStringNumber(w, r)
	if err != nil {
		return
	}

	if !v.AppRepo.DoesAppIdExist(appId) {
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
	doesExist, err := v.VersionRepo.DoesVersionIdExist(versionId)
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

	versionInfo, err := v.VersionRepo.GetVersion(versionId)
	if err != nil {
		u.Logger.Error("error when accessing version info", deepstack.ErrorField, err)
		http.Error(w, "error when accessing version info", http.StatusBadRequest)
		return
	}

	u.SendJsonResponse(w, versionInfo)
}
