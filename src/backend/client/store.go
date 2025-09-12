package store

import (
	"fmt"

	"github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
)

var (
	ApiPrefix    = "/api"
	WipeDataPath = ApiPrefix + "/wipe-data"

	userPath            = ApiPrefix + "/account"
	RegistrationPath    = userPath + "/registration"
	EmailValidationPath = userPath + "/validate"
	LoginPath           = userPath + "/login"
	LogoutPath          = userPath + "/logout"
	AuthCheckPath       = userPath + "/auth-check"
	DeleteUserPath      = userPath + "/delete"
	ChangePasswordPath  = userPath + "/change-password"

	VersionPath       = ApiPrefix + "/versions"
	VersionUploadPath = VersionPath + "/upload"
	VersionDeletePath = VersionPath + "/delete"
	GetVersionsPath   = VersionPath + "/list"
	DownloadPath      = VersionPath + "/download"

	AppPath         = ApiPrefix + "/apps"
	AppCreationPath = AppPath + "/create"
	AppGetListPath  = AppPath + "/get-list"
	AppDeletePath   = AppPath + "/delete"
	SearchAppsPath  = AppPath + "/search"

	DefaultValidationCode = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

type AppStoreClient interface {
	RegisterAndValidateUser(user, password, email string) error
	RegisterUser(user, password, email string) error
	ValidateCode() error
	Login(username, password string) error
	DeleteUser() error
	CreateApp(appName string) (string, error)
	SearchForApps(searchTerm string, showUnofficialApps bool) ([]AppWithLatestVersion, error)
	ListOwnApps() ([]App, error)
	UploadVersion(appId, versionName string, content []byte) (string, error)
	DownloadVersion(versionId string) (*FullVersionInfo, error)
	GetVersions(appId string) ([]Version, error)
	DeleteVersion(versionId string) error
	DeleteApp(appId string) error
	ChangePassword(oldPassword, newPassword string) error
	WipeData()
	Logout() error
	CheckAuth() error
}

type AppStoreClientImpl struct {
	Parent utils.ComponentClient
}

func (h *AppStoreClientImpl) RegisterAndValidateUser(user, password, email string) error {
	err := h.RegisterUser(user, password, email)
	if err != nil {
		return err
	}
	return h.ValidateCode()
}

func (h *AppStoreClientImpl) RegisterUser(user, password, email string) error {
	form := RegistrationForm{
		User:     user,
		Password: password,
		Email:    email,
	}
	_, err := h.Parent.DoRequest(RegistrationPath, form)
	return err
}

func (h *AppStoreClientImpl) ValidateCode() error {
	_, err := h.Parent.DoRequest(EmailValidationPath+"?code="+DefaultValidationCode, nil)
	return err
}

func (h *AppStoreClientImpl) Login(username, password string) error {
	creds := LoginCredentials{
		User:     username,
		Password: password,
	}

	resp, err := h.Parent.DoRequestWithFullResponse(LoginPath, creds)
	if err != nil {
		return err
	}

	cookies := resp.Cookies()
	if len(cookies) != 1 {
		return fmt.Errorf("Expected 1 cookie, got %d", len(cookies))
	}
	h.Parent.Cookie = cookies[0]
	return nil
}

func (h *AppStoreClientImpl) DeleteUser() error {
	_, err := h.Parent.DoRequest(DeleteUserPath, nil)
	return err
}

func (h *AppStoreClientImpl) CreateApp(appName string) (string, error) {
	_, err := h.Parent.DoRequest(AppCreationPath, AppNameString{appName})
	if err != nil {
		return "", err
	}
	appsInStore, err := h.ListOwnApps()
	if err != nil {
		return "", err
	}
	for _, appInStore := range appsInStore {
		if appInStore.Name == appName {
			return appInStore.Id, nil
		}
	}
	return "", fmt.Errorf("app not found on server")
}

func (h *AppStoreClientImpl) SearchForApps(searchTerm string, showUnofficialApps bool) ([]AppWithLatestVersion, error) {
	appSearchRequest := AppSearchRequest{
		SearchTerm:         searchTerm,
		ShowUnofficialApps: showUnofficialApps,
	}
	result, err := h.Parent.DoRequest(SearchAppsPath, appSearchRequest)
	if err != nil {
		return nil, err
	}

	apps, err := utils.UnpackResponse[[]AppWithLatestVersion](result)
	if err != nil {
		return nil, err
	}

	return *apps, nil
}

func (h *AppStoreClientImpl) ListOwnApps() ([]App, error) {
	result, err := h.Parent.DoRequest(AppGetListPath, nil)
	if err != nil {
		return nil, err
	}

	apps, err := utils.UnpackResponse[[]App](result)
	if err != nil {
		return nil, err
	}

	return *apps, nil
}

func (h *AppStoreClientImpl) UploadVersion(appId, versionName string, content []byte) (string, error) {
	tapUpload := &VersionUpload{
		AppId:   appId,
		Version: versionName,
		Content: content,
	}
	_, err := h.Parent.DoRequest(VersionUploadPath, tapUpload)
	if err != nil {
		return "", err
	}

	versionsInStore, err := h.GetVersions(appId)
	if err != nil {
		return "", err
	}
	for _, versionInStore := range versionsInStore {
		if versionInStore.Name == versionName {
			return versionInStore.Id, nil
		}
	}
	return "", fmt.Errorf("version not found on server")
}

func (h *AppStoreClientImpl) DownloadVersion(versionId string) (*FullVersionInfo, error) {
	result, err := h.Parent.DoRequest(DownloadPath, NumberString{versionId})
	if err != nil {
		return nil, err
	}

	fullVersionInfo, err := utils.UnpackResponse[FullVersionInfo](result)
	if err != nil {
		return nil, err
	}

	err = validation.ValidateVersion(fullVersionInfo.Content, fullVersionInfo.Maintainer, fullVersionInfo.AppName)
	if err != nil {
		return nil, fmt.Errorf("version validation failed: %w", err)
	}

	return fullVersionInfo, nil
}

func (h *AppStoreClientImpl) GetVersions(appId string) ([]Version, error) {
	result, err := h.Parent.DoRequest(GetVersionsPath, NumberString{appId})
	if err != nil {
		return nil, err
	}

	versions, err := utils.UnpackResponse[[]Version](result)
	if err != nil {
		return nil, err
	}

	return *versions, nil
}

func (h *AppStoreClientImpl) DeleteVersion(versionId string) error {
	_, err := h.Parent.DoRequest(VersionDeletePath, NumberString{versionId})
	return err
}

func (h *AppStoreClientImpl) DeleteApp(appId string) error {
	_, err := h.Parent.DoRequest(AppDeletePath, NumberString{appId})
	return err
}

func (h *AppStoreClientImpl) ChangePassword(oldPassword, newPassword string) error {
	form := ChangePasswordForm{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	_, err := h.Parent.DoRequest(ChangePasswordPath, form)
	return err
}

func (h *AppStoreClientImpl) WipeData() {
	_, err := h.Parent.DoRequest(WipeDataPath, nil)
	if err != nil {
		panic("failed to wipe data: " + err.Error())
	}
}

func (h *AppStoreClientImpl) Logout() error {
	_, err := h.Parent.DoRequest(LogoutPath, nil)
	return err
}

func (h *AppStoreClientImpl) CheckAuth() error {
	_, err := h.Parent.DoRequest(AuthCheckPath, nil)
	return err
}
