//go:build component

package check

import (
	"net/http"
	"ocelot/store/setup"
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

	// TODO !! also test non-existing maintainer
	_, err := hub.SearchForApps("", "notexistingapp", true)
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
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectUsernameOrPasswordError)
}

func TestLoginSecurity(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	err := hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail)
	assert.Nil(t, err)

	assert.Nil(t, hub.Parent.Cookie)
	assert.Nil(t, hub.Login(tools.SampleUser, tools.SamplePassword))
	assert.NotNil(t, hub.Parent.Cookie)
	// cookie1ExpirationTime := hub.Parent.Cookie.Expires
	checkCookie(t, hub)

	_, err = hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	// cookie2ExpirationTime := hub.Parent.Cookie.Expires
	checkCookie(t, hub)

	// TODO !! implement cookie renewal
	// cookie shall renew its expiration date after authenticated call
	// assert.True(t, cookie1ExpirationTime.Before(cookie2ExpirationTime))

	hub.Parent.Cookie = nil
	correctlyFormattedButNotMatchingPassword := tools.SamplePassword + "x"
	err = hub.Login(tools.SampleUser, correctlyFormattedButNotMatchingPassword)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectUsernameOrPasswordError)
}

func checkCookie(t *testing.T, hub *store.AppStoreClientImpl) {
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
	u.AssertDeepStackErrorFromRequest(t, err, "zip: not a valid zip file")
}

func TestCookieAndHostProtection(t *testing.T) {
	client := GetHubAndLogin(t)
	defer client.WipeData()

	client.Parent.SetCookieHeader = false
	_, err := client.ListOwnApps()
	u.AssertDeepStackErrorFromRequest(t, err, setup.CookieNotSetInRequest)

	client.Parent.SetCookieHeader = true
	client.Parent.Cookie.Value = "some-invalid-cookie-value"
	_, err = client.ListOwnApps()
	u.AssertDeepStackErrorFromRequest(t, err, users.InvalidCookieError)

	validButNonExistentCookie := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	client.Parent.Cookie.Value = validButNonExistentCookie
	_, err = client.ListOwnApps()
	u.AssertDeepStackErrorFromRequest(t, err, users.CookieNotFoundError)

	assert.Nil(t, client.Login(tools.SampleUser, tools.SamplePassword))
}

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
