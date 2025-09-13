//go:build component

package check

import (
	"net/http"
	"ocelot/store/tools"
	"ocelot/store/users"
	"testing"
	"time"

	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
)

var DaysToCookieExpiration = 7

func TestFindAppsSecurity(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	hub.Parent.SetCookieHeader = false

	_, err := hub.SearchForApps("notexistingapp", true)
	assert.Nil(t, err)
}

func TestDownloadAppSecurity(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	versionId, err := hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)

	hub.Parent.SetCookieHeader = false
	fullVersionInfo, err := hub.DownloadVersion(versionId)
	assert.Nil(t, err)
	assert.Equal(t, SampleVersionFileContent, fullVersionInfo.Content)
}

func TestGetVersionsSecurity(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	_, err = hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)

	hub.Parent.SetCookieHeader = false
	versions, err := hub.GetVersions(appId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(versions))
	assert.Equal(t, tools.SampleVersion, versions[0].Name)
}

func TestChangePasswordSecurity(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()

	newPassword := tools.SamplePassword + "x"
	correctlyFormattedButNotMatchingPassword := tools.SamplePassword + "xy"

	err := hub.ChangePassword(correctlyFormattedButNotMatchingPassword, newPassword)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectUsernameAndPasswordError)
}

// TODO test input validation through u.ReadJsonFromRequest

func TestLoginSecurity(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	err := hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail)
	assert.Nil(t, err)

	assert.Nil(t, hub.Parent.Cookie)
	assert.Nil(t, hub.Login(tools.SampleUser, tools.SamplePassword))
	assert.NotNil(t, hub.Parent.Cookie)
	checkCookie(t, hub)

	// cookies are renewed after each successful operation
	_, err = hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	checkCookie(t, hub)

	hub.Parent.Cookie = nil
	correctlyFormattedButNotMatchingPassword := tools.SamplePassword + "x"
	err = hub.Login(tools.SampleUser, correctlyFormattedButNotMatchingPassword)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectUsernameAndPasswordError)
}

func checkCookie(t *testing.T, hub *store.AppStoreClientImpl) {
	assert.Equal(t, "/", hub.Parent.Cookie.Path)
	assert.Equal(t, http.SameSiteStrictMode, hub.Parent.Cookie.SameSite)
	assert.True(t, time.Now().UTC().AddDate(0, 0, DaysToCookieExpiration-1).Before(hub.Parent.Cookie.Expires))
	assert.True(t, time.Now().UTC().AddDate(0, 0, DaysToCookieExpiration+1).After(hub.Parent.Cookie.Expires))
}

func TestUploadOfInvalidZipContent(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	content := []byte("not-bytes-of-valid-zip-file")
	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	_, err = hub.UploadVersion(appId, tools.SampleVersion, content)
	assert.NotNil(t, err)
	// TODO !! can status code context be removed? I think "zip: not a valid zip file" should be the errors message
	u.AssertDeepStackErrorFromRequest(t, err, "zip: not a valid zip file")
}

// TODO !! when integration tests are applied to docker deployment, then there is not need to expose the database port to the host any longer

/* TODO !! -> use two sample endpoints for protection, one for anonymous and one for authenticated users
func TestCookieAndHostProtection(t *testing.T) {
	hub := GetHub()
	tests := []func() error{
		hub.DeleteUser,
		hub.CreateApp,
		hub.DeleteApp,
		hub.UploadVersion,
		hub.DeleteVersion,
		hub.ChangePassword,
		hub.CheckAuth,
	}
	for _, test := range tests {
		doCookieAndHostPolicyChecks(t, hub, test)
	}
}

func doCookieAndHostPolicyChecks(t *testing.T, hub *store.AppStoreClient, operation func() error) {
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterAndValidateUser())
	assert.Nil(t, hub.Login())

	hub.Parent.SetCookieHeader = false

	err := operation()
	assert.NotNil(t, err)
	assert.Equal(t, u.GetErrMsg(400, "cookie not set in request"), err.Error())

	hub.Parent.SetCookieHeader = true
	hub.Parent.Cookie.Value = "some-invalid-cookie-value"
	err = operation()
	assert.NotNil(t, err)
	assert.Equal(t, u.GetErrMsg(400, InvalidCookieError), err.Error())

	validButNonExistentCookie := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	hub.Parent.Cookie.Value = validButNonExistentCookie
	err = operation()
	assert.NotNil(t, err)
	assert.Equal(t, u.GetErrMsg(400, CookieNotFoundError), err.Error())

	assert.Nil(t, hub.Login())

	hub.Parent.User = users.TestUserWithExpiredCookie
	hub.Email = hub.Email + "x"
	assert.Nil(t, hub.RegisterAndValidateUser())
	assert.Nil(t, hub.Login())
	err = operation()
	assert.NotNil(t, err)
	assert.Equal(t, u.GetErrMsg(400, CookieExpiredError), err.Error())
	assert.True(t, time.Now().UTC().After(hub.Parent.Cookie.Expires))
	hub.Parent.User = tools.SampleUser
	hub.Email = tools.SampleEmail
}
*/

func TestCookie(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	assert.Equal(t, tools.CookieName, hub.Parent.Cookie.Name)
	assert.True(t, u.GetTimeInSevenDays().Add(1*time.Second).After(hub.Parent.Cookie.Expires))
	assert.True(t, u.GetTimeInSevenDays().Add(-1*time.Second).Before(hub.Parent.Cookie.Expires))
	assert.Equal(t, 64, len(hub.Parent.Cookie.Value))

	cookie1 := hub.Parent.Cookie
	err := hub.Login(tools.SampleUser, tools.SamplePassword)
	assert.Nil(t, err)
	cookie2 := hub.Parent.Cookie
	assert.NotNil(t, cookie2)
	assert.NotEqual(t, cookie1.Value, cookie2.Value)
}

// TODO !! test cookie renewal on authenticated endpoint calls
