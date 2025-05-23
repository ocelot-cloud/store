package check

import (
	"fmt"
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/utils"
	"ocelot/store/tools"
	"testing"
)

type HubClient struct {
	Parent             utils.ComponentClient
	Email              string
	App                string
	Version            string
	UploadContent      []byte
	AppId              string
	VersionId          string
	ValidationCode     string
	ShowUnofficialApps bool
}

type Operation int

const (
	FindApps Operation = iota
	DownloadVersion
	Register
	ChangePassword
	Login
	CreateApp
	DeleteApp
	UploadVersion
	DeleteVersion
	GetVersions
	CheckAuth
	Validate
)

func getRegistrationForm(hub *HubClient) *tools.RegistrationForm {
	return &tools.RegistrationForm{
		User:     hub.Parent.User,
		Password: hub.Parent.Password,
		Email:    hub.Email,
	}
}

func getHub() *HubClient {
	hub := getHubWithoutWipe()
	hub.wipeData()
	return hub
}

var SampleVersionFileContent = tools.GetValidVersionBytesOfSampleMaintainerApp()

func getHubWithoutWipe() *HubClient {
	return &HubClient{
		Parent: utils.ComponentClient{
			User:            tools.SampleUser,
			Password:        tools.SamplePassword,
			SetCookieHeader: true,
			RootUrl:         tools.RootUrl,
		},

		Email:              tools.SampleEmail,
		App:                tools.SampleApp,
		Version:            tools.SampleVersion,
		UploadContent:      SampleVersionFileContent,
		AppId:              "0",
		VersionId:          "0",
		ValidationCode:     "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		ShowUnofficialApps: true,
	}
}

func (h *HubClient) registerAndValidateUser() error {
	err := h.registerUser()
	if err != nil {
		return err
	}
	return h.validateCode()
}

func (h *HubClient) registerUser() error {
	form := getRegistrationForm(h)
	_, err := h.Parent.DoRequest(tools.RegistrationPath, form, "")
	return err
}

func (h *HubClient) validateCode() error {
	_, err := h.Parent.DoRequest(tools.EmailValidationPath+"?code="+h.ValidationCode, nil, "")
	return err
}

func (h *HubClient) login() error {
	creds := tools.LoginCredentials{
		User:     h.Parent.User,
		Password: h.Parent.Password,
	}

	resp, err := h.Parent.DoRequestWithFullResponse(tools.LoginPath, creds, "")
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

func (h *HubClient) deleteUser() error {
	_, err := h.Parent.DoRequest(tools.DeleteUserPath, nil, "")
	return err
}

func (h *HubClient) createApp() error {
	_, err := h.Parent.DoRequest(tools.AppCreationPath, tools.AppNameString{Value: h.App}, "")
	if err != nil {
		return err
	}
	apps, err := h.ListOwnApps()
	if err != nil {
		return err
	}
	for _, app := range apps {
		if app.Name == h.App {
			h.AppId = app.Id
			return nil
		}
	}
	return fmt.Errorf("app not found on server")
}

func (h *HubClient) SearchForApps(searchTerm string) ([]tools.AppWithLatestVersion, error) {
	appSearchRequest := tools.AppSearchRequest{
		SearchTerm:         searchTerm,
		ShowUnofficialApps: h.ShowUnofficialApps,
	}
	result, err := h.Parent.DoRequest(tools.SearchAppsPath, appSearchRequest, "")
	if err != nil {
		return nil, err
	}

	apps, err := utils.UnpackResponse[[]tools.AppWithLatestVersion](result)
	if err != nil {
		return nil, err
	}

	return *apps, nil
}

func (h *HubClient) ListOwnApps() ([]tools.App, error) {
	result, err := h.Parent.DoRequest(tools.AppGetListPath, nil, "")
	if err != nil {
		return nil, err
	}

	apps, err := utils.UnpackResponse[[]tools.App](result)
	if err != nil {
		return nil, err
	}

	return *apps, nil
}

func (h *HubClient) uploadVersion() error {
	tapUpload := &tools.VersionUpload{
		AppId:   h.AppId,
		Version: h.Version,
		Content: h.UploadContent,
	}
	_, err := h.Parent.DoRequest(tools.VersionUploadPath, tapUpload, "")
	if err != nil {
		return err
	}

	versions, err := h.getVersions()
	if err != nil {
		return err
	}
	for _, version := range versions {
		if version.Name == h.Version {
			h.VersionId = version.Id
			return nil
		}
	}
	return fmt.Errorf("version not found on server")
}

func (h *HubClient) downloadVersion() (*tools.FullVersionInfo, error) {
	result, err := h.Parent.DoRequest(tools.DownloadPath, tools.NumberString{Value: h.VersionId}, "")
	if err != nil {
		return nil, err
	}

	fullVersionInfo, err := utils.UnpackResponse[tools.FullVersionInfo](result)
	if err != nil {
		return nil, err
	}

	return fullVersionInfo, nil
}

func (h *HubClient) getVersions() ([]tools.Version, error) {
	result, err := h.Parent.DoRequest(tools.GetVersionsPath, tools.NumberString{Value: h.AppId}, "")
	if err != nil {
		return nil, err
	}

	versions, err := utils.UnpackResponse[[]tools.Version](result)
	if err != nil {
		return nil, err
	}

	return *versions, nil
}

func (h *HubClient) deleteVersion() error {
	_, err := h.Parent.DoRequest(tools.VersionDeletePath, tools.NumberString{Value: h.VersionId}, "")
	return err
}

func (h *HubClient) deleteApp() error {
	_, err := h.Parent.DoRequest(tools.AppDeletePath, tools.NumberString{Value: h.AppId}, "")
	return err
}

func (h *HubClient) changePassword() error {
	form := tools.ChangePasswordForm{
		OldPassword: h.Parent.Password,
		NewPassword: h.Parent.NewPassword,
	}

	_, err := h.Parent.DoRequest(tools.ChangePasswordPath, form, "")
	return err
}

func getHubAndLogin(t *testing.T) *HubClient {
	hub := getHub()
	assert.Nil(t, hub.registerUser())
	assert.Nil(t, hub.validateCode())
	err := hub.login()
	assert.Nil(t, err)
	return hub
}

func (h *HubClient) wipeData() {
	_, err := h.Parent.DoRequest(tools.WipeDataPath, nil, "")
	if err != nil {
		panic("failed to wipe data: " + err.Error())
	}
}

func (h *HubClient) logout() error {
	_, err := h.Parent.DoRequest(tools.LogoutPath, nil, "")
	return err
}

func (h *HubClient) checkAuth() error {
	_, err := h.Parent.DoRequest(tools.AuthCheckPath, nil, "")
	return err
}
