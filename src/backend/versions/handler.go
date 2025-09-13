package versions

import (
	"encoding/json"
	"errors"
	"net/http"
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"

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

func (v *VersionsHandler) VersionUploadHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	r.Body = http.MaxBytesReader(w, r.Body, tools.MaxPayloadSizeInBytes)
	defer u.Close(r.Body)
	var versionUpload store.VersionUploadDto
	err := json.NewDecoder(r.Body).Decode(&versionUpload)
	if errors.Is(err, errors.New("http: request body too large")) {
		u.Logger.Info("version upload version content of user was too large", tools.UserField, user)
		// TODO !! the "1" should be taken from a global variable
		http.Error(w, "version content too large, the limit is 1MB", http.StatusBadRequest)
		return
	}
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	err = v.VersionService.UploadVersion(user, &versionUpload)
	if err != nil {
		// TODO !! space use case to be covered by component tests I guess? also NotOwningThisVersionError
		expectedErrors := u.MapOf(users.InvalidInputError, users.NotEnoughSpacePrefix, NotOwningThisVersionError, NotAZipFileError, VersionAlreadyExist, AppDoesNotExist)
		u.WriteResponseError(w, expectedErrors, err)
		return
	}
}

func (v *VersionsHandler) VersionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	versionId, ok := apps.ReadBodyAsStringNumber(w, r)
	if !ok {
		return
	}
	err := v.VersionService.DeleteVersionWithChecks(user.Id, versionId)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(NotOwningThisVersionError, VersionDoesNotExistError), err, tools.UserField, user, tools.VersionIdField, versionId)
	}
}

func (v *VersionsHandler) GetVersionsHandler(w http.ResponseWriter, r *http.Request) {
	appId, ok := apps.ReadBodyAsStringNumber(w, r)
	if !ok {
		return
	}
	versionsList, err := v.VersionService.ListVersions(appId)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(AppDoesNotExist), err)
		return
	}
	u.SendJsonResponse(w, versionsList)
}

func (v *VersionsHandler) VersionDownloadHandler(w http.ResponseWriter, r *http.Request) {
	versionId, ok := apps.ReadBodyAsStringNumber(w, r)
	if !ok {
		return
	}
	versionInfo, err := v.VersionService.GetVersionForDownload(versionId)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(VersionDoesNotExistError), err)
		return
	}
	u.SendJsonResponse(w, versionInfo)
}
