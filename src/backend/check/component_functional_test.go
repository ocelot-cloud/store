//go:build component

package check

import (
	"ocelot/store/tools"
	"testing"
	"time"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/utils"
)

func TestVersionDownload(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail))
	assert.Nil(t, hub.Login(tools.SampleUser, tools.SamplePassword))

	notExistingVersionId := "0"
	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	_, err = hub.DownloadVersion(notExistingVersionId)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "version does not exist", 404)

	versionId, err := hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)
	foundVersions, err := hub.GetVersions(appId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundVersions))
	assert.Equal(t, tools.SampleVersion, foundVersions[0].Name)

	fullVersionInfo, err := hub.DownloadVersion(versionId)
	assert.Nil(t, err)
	assert.Equal(t, SampleVersionFileContent, fullVersionInfo.Content)
}

func TestCookie(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	assert.Equal(t, tools.CookieName, hub.Parent.Cookie.Name)
	assert.True(t, utils.GetTimeInSevenDays().Add(1*time.Second).After(hub.Parent.Cookie.Expires))
	assert.True(t, utils.GetTimeInSevenDays().Add(-1*time.Second).Before(hub.Parent.Cookie.Expires))
	assert.Equal(t, 64, len(hub.Parent.Cookie.Value))

	cookie1 := hub.Parent.Cookie
	err := hub.Login(tools.SampleUser, tools.SamplePassword)
	assert.Nil(t, err)
	cookie2 := hub.Parent.Cookie
	assert.NotNil(t, cookie2)
	assert.NotEqual(t, cookie1.Value, cookie2.Value)
}

func TestCreateApp(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	foundApps, err := hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundApps))
	foundApp := foundApps[0]
	assert.Equal(t, tools.SampleUser, foundApp.Maintainer)
	assert.Equal(t, tools.SampleApp, foundApp.Name)

	_, err = hub.CreateApp(tools.SampleApp)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "app already exists", 409)

	assert.Nil(t, hub.DeleteApp(appId))
	foundApps, err = hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundApps))
}

func TestUploadVersion(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	notExistingVersionId := "0"
	_, err := hub.UploadVersion(notExistingVersionId, tools.SampleVersion, SampleVersionFileContent)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "app does not exist", 404)

	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	versionId, err := hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)

	_, err = hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "version already exists", 409)

	versions, err := hub.GetVersions(appId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(versions))
	assert.Equal(t, tools.SampleVersion, versions[0].Name)

	assert.Nil(t, hub.DeleteVersion(versionId))
	versions, err = hub.GetVersions(appId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(versions))

	err = hub.DeleteVersion(versionId)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "version does not exist", 404)
}

func TestLogin(t *testing.T) {
	hub := GetHub()
	err := hub.Login(tools.SampleUser, tools.SamplePassword)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "user does not exist", 404)
}

func TestChangePassword(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	newPassword := tools.SamplePassword + "x"

	assert.Nil(t, hub.ChangePassword(tools.SamplePassword, newPassword))
	err := hub.Login(tools.SampleUser, tools.SamplePassword)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "incorrect username or password", 401)

	hub.Parent.Cookie = nil
	err = hub.Login(tools.SampleUser, newPassword)
	assert.Nil(t, err)
	assert.NotNil(t, hub.Parent.Cookie)
}

func TestRegistration(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail))
	err := hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "user already exists", 409)
}

func TestGetVersionsUnhappyPath(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail))
	assert.Nil(t, hub.Login(tools.SampleUser, tools.SamplePassword))

	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	versionList, err := hub.GetVersions(appId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(versionList))
}

func TestLogout(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	_, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	assert.Nil(t, hub.Logout())
	_, err = hub.CreateApp(tools.SampleApp)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "cookie not found", 401)
}

func TestGetAppList(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	apps, err := hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(apps))
	_, err = hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	apps, err = hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, tools.SampleApp, apps[0].Name)
}

func TestRegistrationAndValidation(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail))
	assert.NotNil(t, hub.Login(tools.SampleUser, tools.SamplePassword))
	assert.Nil(t, hub.ValidateCode())
	assert.Nil(t, hub.Login(tools.SampleUser, tools.SamplePassword))
}

func TestEmailAlreadyExists(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail))
	user2 := tools.SampleUser + "2"
	err := hub.RegisterUser(user2, tools.SamplePassword, tools.SampleEmail)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "email already exists", 409)
}

func TestDownloadDummyVersion(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	versionId, err := hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)
	apps, err := hub.SearchForApps(tools.SampleApp, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	versions, err := hub.GetVersions(appId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(versions))

	info, err := hub.DownloadVersion(versionId)
	assert.Nil(t, err)
	assert.Equal(t, tools.SampleUser, info.Maintainer)
	assert.Equal(t, tools.SampleApp, info.AppName)
	assert.Equal(t, tools.SampleVersion, info.VersionName)
	assert.True(t, len(info.Content) > 100)
}

func TestCreationOfOcelotCloudAppIsForbidden(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	_, err := hub.CreateApp("ocelotcloud")
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "app name is reserved", 400)
}

func TestUnofficialAppFilteringWhenSearching(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	_, err = hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)

	apps, err := hub.SearchForApps(tools.SampleApp, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, appId, apps[0].AppId)

	apps, err = hub.SearchForApps(tools.SampleApp, false)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(apps))
}

func TestAllowEmptyStringAsSearchTerm(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	apps, err := hub.SearchForApps("", true)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(apps))

	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	_, err = hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)

	apps, err = hub.SearchForApps("", true)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
}

// TODO !! duplication with cloud -> move to "shared"
func AssertDeepStackErrorWithCode(t *testing.T, err error, expectedResponseBodyErrorMessage string, expectedStatusCode int) {
	asd
	deepstack.AssertDeepStackError(t, err, "request failed", "response_body", expectedResponseBodyErrorMessage, "status_code", expectedStatusCode)
}
