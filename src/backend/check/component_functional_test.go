//go:build component

package check

import (
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/utils"
	"ocelot/store/tools"
	"testing"
	"time"
)

func TestVersionDownload(t *testing.T) {
	hub := GetHub()

	_, err := hub.DownloadVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "version does not exist"), err.Error())

	assert.Nil(t, hub.RegisterAndValidateUser())
	assert.Nil(t, hub.Login())

	_, err = hub.DownloadVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "version does not exist"), err.Error())

	assert.Nil(t, hub.CreateApp())
	_, err = hub.DownloadVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "version does not exist"), err.Error())

	assert.Nil(t, hub.UploadVersion())
	foundVersions, err := hub.GetVersions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundVersions))
	assert.Equal(t, tools.SampleVersion, foundVersions[0].Name)

	fullVersionInfo, err := hub.DownloadVersion()
	assert.Nil(t, err)
	assert.Equal(t, SampleVersionFileContent, fullVersionInfo.Content)
}

func TestCookie(t *testing.T) {
	hub := GetHubAndLogin(t)
	assert.Equal(t, tools.CookieName, hub.Parent.Cookie.Name)
	assert.True(t, utils.GetTimeInSevenDays().Add(1*time.Second).After(hub.Parent.Cookie.Expires))
	assert.True(t, utils.GetTimeInSevenDays().Add(-1*time.Second).Before(hub.Parent.Cookie.Expires))
	assert.Equal(t, 64, len(hub.Parent.Cookie.Value))

	cookie1 := hub.Parent.Cookie
	err := hub.Login()
	assert.Nil(t, err)
	cookie2 := hub.Parent.Cookie
	assert.NotNil(t, cookie2)
	assert.NotEqual(t, cookie1.Value, cookie2.Value)
}

func TestCreateApp(t *testing.T) {
	hub := GetHubAndLogin(t)
	assert.Nil(t, hub.CreateApp())
	foundApps, err := hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundApps))
	foundApp := foundApps[0]
	assert.Equal(t, hub.Parent.User, foundApp.Maintainer)
	assert.Equal(t, hub.App, foundApp.Name)

	err = hub.CreateApp()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(409, "app already exists"), err.Error())

	assert.Nil(t, hub.DeleteApp())
	foundApps, err = hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundApps))
}

func TestUploadVersion(t *testing.T) {
	hub := GetHubAndLogin(t)

	err := hub.UploadVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "app does not exist"), err.Error())

	assert.Nil(t, hub.CreateApp())
	assert.Nil(t, hub.UploadVersion())

	err = hub.UploadVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(409, "version already exists"), err.Error())

	versions, err := hub.GetVersions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(versions))
	assert.Equal(t, hub.Version, versions[0].Name)

	assert.Nil(t, hub.DeleteVersion())
	versions, err = hub.GetVersions()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(versions))

	err = hub.DeleteVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "version does not exist"), err.Error())
}

func TestLogin(t *testing.T) {
	hub := GetHub()
	err := hub.Login()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "user does not exist"), err.Error())
}

func TestChangePassword(t *testing.T) {
	hub := GetHubAndLogin(t)

	hub.Parent.NewPassword = hub.Parent.Password + "x"

	assert.Nil(t, hub.ChangePassword())
	err := hub.Login()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(401, "incorrect username or password"), err.Error())

	hub.Parent.Password = hub.Parent.NewPassword
	hub.Parent.Cookie = nil
	err = hub.Login()
	assert.Nil(t, err)
	assert.NotNil(t, hub.Parent.Cookie)
}

func TestRegistration(t *testing.T) {
	hub := GetHub()
	assert.Nil(t, hub.RegisterAndValidateUser())
	err := hub.RegisterAndValidateUser()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(409, "user already exists"), err.Error())
}

func TestGetVersionsUnhappyPath(t *testing.T) {
	hub := GetHub()

	assert.Nil(t, hub.RegisterAndValidateUser())
	assert.Nil(t, hub.Login())
	_, err := hub.GetVersions()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "app does not exist"), err.Error())

	assert.Nil(t, hub.CreateApp())
	versionList, err := hub.GetVersions()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(versionList))
}

func TestLogout(t *testing.T) {
	hub := GetHubAndLogin(t)
	assert.Nil(t, hub.CreateApp())
	assert.Nil(t, hub.Logout())
	err := hub.CreateApp()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(401, "cookie not found"), err.Error())
}

func TestGetAppList(t *testing.T) {
	hub := GetHubAndLogin(t)
	apps, err := hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(apps))
	assert.Nil(t, hub.CreateApp())
	apps, err = hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, tools.SampleApp, apps[0].Name)
}

func TestRegistrationAndValidation(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterUser())
	assert.NotNil(t, hub.Login())
	assert.Nil(t, hub.ValidateCode())
	assert.Nil(t, hub.Login())
}

func TestEmailAlreadyExists(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterAndValidateUser())
	hub.Parent.User = tools.SampleUser + "2"
	err := hub.RegisterUser()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(409, "email already exists"), err.Error())
}

func TestDownloadDummyVersion(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	assert.Nil(t, hub.CreateApp())
	assert.Nil(t, hub.UploadVersion())
	apps, err := hub.SearchForApps(hub.App)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	hub.AppId = apps[0].AppId
	versions, err := hub.GetVersions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(versions))
	hub.VersionId = versions[0].Id

	info, err := hub.DownloadVersion()
	assert.Nil(t, err)
	assert.Equal(t, hub.Parent.User, info.Maintainer)
	assert.Equal(t, hub.App, info.AppName)
	assert.Equal(t, hub.Version, info.VersionName)
	assert.True(t, len(info.Content) > 100)
}

func TestCreationOfOcelotCloudAppIsForbidden(t *testing.T) {
	hub := GetHubAndLogin(t)
	hub.App = "ocelotcloud"
	err := hub.CreateApp()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(400, "app name is reserved"), err.Error())
}

func TestUnofficialAppFilteringWhenSearching(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	assert.Nil(t, hub.CreateApp())
	assert.Nil(t, hub.UploadVersion())

	hub.ShowUnofficialApps = true
	apps, err := hub.SearchForApps(hub.App)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, hub.AppId, apps[0].AppId)

	hub.ShowUnofficialApps = false
	apps, err = hub.SearchForApps(hub.App)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(apps))
}

func TestAllowEmptyStringAsSearchTerm(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	apps, err := hub.SearchForApps("")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(apps))

	assert.Nil(t, hub.CreateApp())
	assert.Nil(t, hub.UploadVersion())

	apps, err = hub.SearchForApps("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))

}
