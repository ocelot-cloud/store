//go:build component

package check

import (
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/utils"
	"net/http"
	"ocelot/store/tools"
	"ocelot/store/users"
	"testing"
	"time"
)

var DaysToCookieExpiration = 7

func TestCorsHeaderArePresentInTestProfile(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	response, err := hub.Parent.DoRequestWithFullResponse(tools.AppGetListPath, nil)
	assert.Nil(t, err)

	assert.Equal(t, "", response.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", response.Header.Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "POST, GET, OPTIONS, PUT, DELETE", response.Header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Accept, Content-Type, Content-Length, Authorization", response.Header.Get("Access-Control-Allow-Headers"))
}

func TestFindAppsSecurity(t *testing.T) {
	hub := GetHubAndLogin(t)
	hub.Parent.SetCookieHeader = false

	_, err := hub.SearchForApps("notexistingapp")
	assert.Nil(t, err)

	testInputInvalidation(t, hub, "not-existing-app", SearchTerm, FindApps)
}

func TestDownloadAppSecurity(t *testing.T) {
	hub := GetHubAndLogin(t)

	assert.Nil(t, hub.CreateApp())
	assert.Nil(t, hub.UploadVersion())

	hub.Parent.SetCookieHeader = false
	fullVersionInfo, err := hub.DownloadVersion()
	assert.Nil(t, err)
	assert.Equal(t, SampleVersionFileContent, fullVersionInfo.Content)
}

func TestGetVersionsSecurity(t *testing.T) {
	hub := GetHubAndLogin(t)

	assert.Nil(t, hub.CreateApp())
	assert.Nil(t, hub.UploadVersion())

	hub.Parent.SetCookieHeader = false
	versions, err := hub.GetVersions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(versions))
	assert.Equal(t, tools.SampleVersion, versions[0].Name)
}

func TestRegisterSecurity(t *testing.T) {
	hub := GetHub()
	hub.Parent.SetCookieHeader = false
	testInputInvalidation(t, hub, "invalid-password-with-letter-ä", PasswordField, Register)
	testInputInvalidation(t, hub, "invalid-username", UserField, Register)
}

func TestChangePasswordSecurity(t *testing.T) {
	hub := GetHubAndLogin(t)

	hub.Parent.NewPassword = tools.SamplePassword + "x"
	correctlyFormattedButNotMatchingPassword := tools.SamplePassword + "xy"
	hub.Parent.Password = correctlyFormattedButNotMatchingPassword
	err := hub.ChangePassword()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(401, "incorrect username or password"), err.Error())
	hub.Parent.Password = tools.SamplePassword

	testInputInvalidation(t, hub, "invalid-password-ä", PasswordField, ChangePassword)
	testInputInvalidation(t, hub, "invalid-password-ä", NewPasswordField, ChangePassword)
}

func TestLoginSecurity(t *testing.T) {
	hub := GetHub()
	err := hub.RegisterAndValidateUser()
	assert.Nil(t, err)

	assert.Nil(t, hub.Parent.Cookie)
	assert.Nil(t, hub.Login())
	assert.NotNil(t, hub.Parent.Cookie)
	checkCookie(t, hub)

	// cookies are renewed after each successful operation
	assert.Nil(t, hub.CreateApp())
	checkCookie(t, hub)

	hub.Parent.Cookie = nil
	testInputInvalidation(t, hub, "invalid-user", UserField, Login)
	testInputInvalidation(t, hub, "invalid-password-ä", PasswordField, Login)

	correctlyFormattedButNotMatchingPassword := tools.SamplePassword + "x"
	hub.Parent.Password = correctlyFormattedButNotMatchingPassword
	err = hub.Login()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(401, "incorrect username or password"), err.Error())
	hub.Parent.Password = tools.SamplePassword
}

func checkCookie(t *testing.T, hub *AppStoreClient) {
	assert.Equal(t, "/", hub.Parent.Cookie.Path)
	assert.Equal(t, http.SameSiteStrictMode, hub.Parent.Cookie.SameSite)
	assert.True(t, time.Now().UTC().AddDate(0, 0, DaysToCookieExpiration-1).Before(hub.Parent.Cookie.Expires))
	assert.True(t, time.Now().UTC().AddDate(0, 0, DaysToCookieExpiration+1).After(hub.Parent.Cookie.Expires))
}

func TestCreateAppSecurity(t *testing.T) {
	hub := GetHubAndLogin(t)
	testInputInvalidation(t, hub, "invalid_app", AppField, CreateApp)
}

func TestUploadVersionSecurity(t *testing.T) {
	hub := GetHubAndLogin(t)
	testInputInvalidation(t, hub, "invalid-version", VersionField, UploadVersion)
}

func TestCookieExpirationAndRenewal(t *testing.T) {
	hub := GetHubAndLogin(t)
	// There is some specific logic for this user in the production code when handling cookie.
	hub.Parent.User = users.TestUserWithExpiredCookie
	hub.Email = hub.Email + "x"
	assert.Nil(t, hub.RegisterAndValidateUser())
	assert.Nil(t, hub.Login())
	assert.True(t, time.Now().UTC().After(hub.Parent.Cookie.Expires))
	err := hub.CreateApp()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(400, "cookie expired"), err.Error())
	hub.Parent.User = tools.SampleUser

	// There is some specific logic for this user in the production code when handling cookie.
	hub.Parent.User = users.TestUserWithOldButNotExpiredCookie
	hub.Email = hub.Email + "y"
	assert.Nil(t, hub.RegisterAndValidateUser())
	assert.Nil(t, hub.Login())
	assert.True(t, time.Now().UTC().Before(hub.Parent.Cookie.Expires))
	assert.True(t, time.Now().UTC().Add(48*time.Hour).After(hub.Parent.Cookie.Expires))
	assert.Nil(t, hub.CreateApp())
	assert.True(t, time.Now().UTC().AddDate(0, 0, DaysToCookieExpiration-1).Before(hub.Parent.Cookie.Expires))
	assert.True(t, time.Now().UTC().AddDate(0, 0, DaysToCookieExpiration+1).After(hub.Parent.Cookie.Expires))
	hub.Parent.User = tools.SampleUser
}

func TestOwnership(t *testing.T) {
	hub := GetHub()
	testVersionOwnership(t, hub, hub.DeleteApp)
	hub = GetHub()
	testVersionOwnership(t, hub, hub.UploadVersion)
}

func testVersionOwnership(t *testing.T, hub *AppStoreClient, operation func() error) {
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterAndValidateUser())
	assert.Nil(t, hub.Login())
	assert.Nil(t, hub.CreateApp())
	hub.Parent.User = tools.SampleUser + "2"
	hub.Email = tools.SampleEmail + "x"
	assert.Nil(t, hub.RegisterAndValidateUser())
	assert.Nil(t, hub.Login())
	err := operation()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(401, "you do not own this app"), err.Error())
}

func TestOwnershipOfDeleteVersion(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	assert.Nil(t, hub.CreateApp())
	assert.Nil(t, hub.UploadVersion())

	hub.Parent.User = tools.SampleUser + "2"
	hub.Email = tools.SampleEmail + "x"
	assert.Nil(t, hub.RegisterAndValidateUser())
	assert.Nil(t, hub.Login())

	err := hub.DeleteVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(401, "you do not own this version"), err.Error())
}

func TestValidationCodeInputValidation(t *testing.T) {
	hub := GetHub()
	testInputInvalidation(t, hub, "?123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", ValidationCodeField, Validate)
	testInputInvalidation(t, hub, "1234", ValidationCodeField, Validate)
}

func TestEmailInputValidation(t *testing.T) {
	hub := GetHub()
	testInputInvalidation(t, hub, "admin@admin", EmailField, Register)
}

func TestIdInputValidationDuringDownload(t *testing.T) {
	hub := GetHubAndLogin(t)
	testInputInvalidation(t, hub, "1234a", VersionIdField, DownloadVersion)
	testInputInvalidation(t, hub, "1234a", VersionIdField, DeleteVersion)
	testInputInvalidation(t, hub, "1234a", AppIdField, GetVersions)
	testInputInvalidation(t, hub, "1234a", AppIdField, UploadVersion)
	testInputInvalidation(t, hub, "1234a", AppIdField, DeleteApp)
}

func TestUploadOfInvalidZipContent(t *testing.T) {
	hub := GetHubAndLogin(t)
	hub.UploadContent = []byte("not-bytes-of-valid-zip-file")
	assert.Nil(t, hub.CreateApp())
	err := hub.UploadVersion()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(400, "invalid version: failed to read zip file: zip: not a valid zip file"), err.Error())
}

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

func doCookieAndHostPolicyChecks(t *testing.T, hub *AppStoreClient, operation func() error) {
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterAndValidateUser())
	assert.Nil(t, hub.Login())

	hub.Parent.SetCookieHeader = false

	err := operation()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(401, "cookie not set in request"), err.Error())

	hub.Parent.SetCookieHeader = true
	hub.Parent.Cookie.Value = "some-invalid-cookie-value"
	err = operation()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(400, "invalid cookie"), err.Error())

	validButNonExistentCookie := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	hub.Parent.Cookie.Value = validButNonExistentCookie
	err = operation()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(401, "cookie not found"), err.Error())

	assert.Nil(t, hub.Login())

	hub.Parent.User = users.TestUserWithExpiredCookie
	hub.Email = hub.Email + "x"
	assert.Nil(t, hub.RegisterAndValidateUser())
	assert.Nil(t, hub.Login())
	err = operation()
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(400, "cookie expired"), err.Error())
	assert.True(t, time.Now().UTC().After(hub.Parent.Cookie.Expires))
	hub.Parent.User = tools.SampleUser
	hub.Email = tools.SampleEmail
}

type FieldType int

const (
	UserField FieldType = iota
	PasswordField
	NewPasswordField
	EmailField
	AppField
	VersionField
	ValidationCodeField
	AppIdField
	VersionIdField
	SearchTerm
)

func testInputInvalidation(t *testing.T, hub *AppStoreClient, invalidValue string, fieldType FieldType, operation Operation) {
	originalValue := returnCurrentValueAndSetField(hub, fieldType, invalidValue)

	switch operation {
	case Register:
		assertInvalidInputError(t, hub.RegisterAndValidateUser())
	case GetVersions:
		_, err := hub.GetVersions()
		assertInvalidInputError(t, err)
	case DownloadVersion:
		_, err := hub.DownloadVersion()
		assertInvalidInputError(t, err)
	case FindApps:
		_, err := hub.SearchForApps(invalidValue)
		assertInvalidInputError(t, err)
	case ChangePassword:
		assertInvalidInputError(t, hub.ChangePassword())
	case Login:
		assertInvalidInputError(t, hub.Login())
	case DeleteApp:
		assertInvalidInputError(t, hub.DeleteApp())
	case UploadVersion:
		assertInvalidInputError(t, hub.UploadVersion())
	case DeleteVersion:
		assertInvalidInputError(t, hub.DeleteVersion())
	case CheckAuth:
		assertInvalidInputError(t, hub.CheckAuth())
	case CreateApp:
		assertInvalidInputError(t, hub.CreateApp())
	case Validate:
		assertInvalidInputError(t, hub.ValidateCode())
	default:
		panic("Unsupported operation")
	}

	returnCurrentValueAndSetField(hub, fieldType, originalValue)
}

func assertInvalidInputError(t *testing.T, err error) {
	assert.NotNil(t, err)
	assert.Equal(t, utils.GetErrMsg(400, "invalid input"), err.Error())
}

func returnCurrentValueAndSetField(hub *AppStoreClient, fieldType FieldType, value string) string {
	var originalValue string
	switch fieldType {
	case PasswordField:
		originalValue = hub.Parent.Password
		hub.Parent.Password = value
	case NewPasswordField:
		originalValue = hub.Parent.NewPassword
		hub.Parent.NewPassword = value
	case UserField:
		originalValue = hub.Parent.User
		hub.Parent.User = value
	case EmailField:
		originalValue = hub.Email
		hub.Email = value
	case AppField:
		originalValue = hub.App
		hub.App = value
	case VersionField:
		originalValue = hub.Version
		hub.Version = value
	case ValidationCodeField:
		originalValue = hub.ValidationCode
		hub.ValidationCode = value
	case VersionIdField:
		originalValue = hub.VersionId
		hub.VersionId = value
	case AppIdField:
		originalValue = hub.AppId
		hub.AppId = value
	case SearchTerm:
	default:
		panic("Unsupported field type")
	}
	return originalValue
}
