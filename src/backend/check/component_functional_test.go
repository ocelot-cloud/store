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
	hub := getHub()

	_, err := hub.downloadVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "version does not exist"), err.Error())

	assert.Nil(t, hub.registerAndValidateUser())
	assert.Nil(t, hub.login())

	_, err = hub.downloadVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "version does not exist"), err.Error())

	assert.Nil(t, hub.createApp())
	_, err = hub.downloadVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "version does not exist"), err.Error())

	assert.Nil(t, hub.uploadVersion())
	foundVersions, err := hub.getVersions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundVersions))
	assert.Equal(t, tools.SampleVersion, foundVersions[0].Name)

	fullVersionInfo, err := hub.downloadVersion()
	assert.Nil(t, err)
	assert.Equal(t, SampleVersionFileContent, fullVersionInfo.Content)
}

func TestCookie(t *testing.T) {
	hub := getHubAndLogin(t)
	assert.Equal(t, tools.CookieName, hub.Parent.Cookie.Name)
	assert.True(t, utils.GetTimeIn30Days().Add(1*time.Second).After(hub.Parent.Cookie.Expires))
	assert.True(t, utils.GetTimeIn30Days().Add(-1*time.Second).Before(hub.Parent.Cookie.Expires))
	assert.Equal(t, 64, len(hub.Parent.Cookie.Value))

	cookie1 := hub.Parent.Cookie
	err := hub.login()
	assert.Nil(t, err)
	cookie2 := hub.Parent.Cookie
	assert.NotNil(t, cookie2)
	assert.NotEqual(t, cookie1.Value, cookie2.Value)
}

func TestCreateApp(t *testing.T) {
	hub := getHubAndLogin(t)
	assert.Nil(t, hub.createApp())
	foundApps, err := hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundApps))
	foundApp := foundApps[0]
	assert.Equal(t, hub.Parent.User, foundApp.Maintainer)
	assert.Equal(t, hub.App, foundApp.Name)

	err = hub.createApp()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(409, "app already exists"), err.Error())

	assert.Nil(t, hub.deleteApp())
	foundApps, err = hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundApps))
}

func TestUploadVersion(t *testing.T) {
	hub := getHubAndLogin(t)

	err := hub.uploadVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "app does not exist"), err.Error())

	assert.Nil(t, hub.createApp())
	assert.Nil(t, hub.uploadVersion())

	err = hub.uploadVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(409, "version already exists"), err.Error())

	versions, err := hub.getVersions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(versions))
	assert.Equal(t, hub.Version, versions[0].Name)

	assert.Nil(t, hub.deleteVersion())
	versions, err = hub.getVersions()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(versions))

	err = hub.deleteVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "version does not exist"), err.Error())
}

func TestLogin(t *testing.T) {
	hub := getHub()
	err := hub.login()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "user does not exist"), err.Error())
}

func TestChangePassword(t *testing.T) {
	hub := getHubAndLogin(t)

	hub.Parent.NewPassword = hub.Parent.Password + "x"

	assert.Nil(t, hub.changePassword())
	err := hub.login()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(401, "incorrect username or password"), err.Error())

	hub.Parent.Password = hub.Parent.NewPassword
	hub.Parent.Cookie = nil
	err = hub.login()
	assert.Nil(t, err)
	assert.NotNil(t, hub.Parent.Cookie)
}

func TestRegistration(t *testing.T) {
	hub := getHub()
	assert.Nil(t, hub.registerAndValidateUser())
	err := hub.registerAndValidateUser()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(409, "user already exists"), err.Error())
}

func TestGetVersionsUnhappyPath(t *testing.T) {
	hub := getHub()

	assert.Nil(t, hub.registerAndValidateUser())
	assert.Nil(t, hub.login())
	_, err := hub.getVersions()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(404, "app does not exist"), err.Error())

	assert.Nil(t, hub.createApp())
	versionList, err := hub.getVersions()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(versionList))
}

func TestLogout(t *testing.T) {
	hub := getHubAndLogin(t)
	assert.Nil(t, hub.createApp())
	assert.Nil(t, hub.logout())
	err := hub.createApp()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(401, "cookie not found"), err.Error())
}

func TestGetAppList(t *testing.T) {
	hub := getHubAndLogin(t)
	apps, err := hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(apps))
	assert.Nil(t, hub.createApp())
	apps, err = hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, tools.SampleApp, apps[0].Name)
}

func TestRegistrationAndValidation(t *testing.T) {
	hub := getHub()
	defer hub.wipeData()
	assert.Nil(t, hub.registerUser())
	assert.NotNil(t, hub.login())
	assert.Nil(t, hub.validateCode())
	assert.Nil(t, hub.login())
}

func TestEmailAlreadyExists(t *testing.T) {
	hub := getHub()
	defer hub.wipeData()
	assert.Nil(t, hub.registerAndValidateUser())
	hub.Parent.User = tools.SampleUser + "2"
	err := hub.registerUser()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(409, "email already exists"), err.Error())
}

func TestDownloadDummyVersion(t *testing.T) {
	hub := getHubAndLogin(t)
	defer hub.wipeData()
	assert.Nil(t, hub.createApp())
	assert.Nil(t, hub.uploadVersion())
	apps, err := hub.SearchForApps(hub.App)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	hub.AppId = apps[0].AppId
	versions, err := hub.getVersions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(versions))
	hub.VersionId = versions[0].Id

	info, err := hub.downloadVersion()
	assert.Nil(t, err)
	assert.Equal(t, hub.Parent.User, info.Maintainer)
	assert.Equal(t, hub.App, info.AppName)
	assert.Equal(t, hub.Version, info.VersionName)
	assert.True(t, len(info.Content) > 100)
}

func TestCreationOfOcelotCloudAppIsForbidden(t *testing.T) {
	hub := getHubAndLogin(t)
	hub.App = "ocelotcloud"
	err := hub.createApp()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(400, "app name is reserved"), err.Error())
}

func TestUnofficialAppFilteringWhenSearching(t *testing.T) {
	hub := getHubAndLogin(t)
	defer hub.wipeData()
	assert.Nil(t, hub.createApp())
	assert.Nil(t, hub.uploadVersion())

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
	hub := getHubAndLogin(t)
	defer hub.wipeData()
	apps, err := hub.SearchForApps("")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(apps))

	assert.Nil(t, hub.createApp())
	assert.Nil(t, hub.uploadVersion())

	apps, err = hub.SearchForApps("")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))

}
