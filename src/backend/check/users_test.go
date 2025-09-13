//go:build component

package check

import (
	"ocelot/store/tools"
	"ocelot/store/users"
	"testing"

	"github.com/ocelot-cloud/shared/assert"
	u "github.com/ocelot-cloud/shared/utils"
)

func TestLogin(t *testing.T) {
	hub := GetHub()
	err := hub.Login(tools.SampleUser, tools.SamplePassword)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "user does not exist")
}

func TestChangePassword(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	newPassword := tools.SamplePassword + "x"

	assert.Nil(t, hub.ChangePassword(tools.SamplePassword, newPassword))
	err := hub.Login(tools.SampleUser, tools.SamplePassword)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.IncorrectUsernameOrPasswordError)

	hub.Parent.Cookie = nil
	err = hub.Login(tools.SampleUser, newPassword)
	assert.Nil(t, err)
	assert.NotNil(t, hub.Parent.Cookie)
}

func TestCanNotCreateUserWhoseNameOrEmailIsAlreadyUsed(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail))

	err := hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail+"x")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.UserAlreadyExistsError)

	err = hub.RegisterAndValidateUser(tools.SampleUser+"x", tools.SamplePassword, tools.SampleEmail)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.EmailAlreadyExistsError)
}

func TestLogout(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	_, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	assert.Nil(t, hub.Logout())
	_, err = hub.CreateApp(tools.SampleApp)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, users.CookieNotFoundError)
}

// TODO !! cant register twice with same 1) name or 2) email
// TODO !! empty or wrong registration code should not be accepted
/* TODO !! assert validation code?
assert.Equal(t, 64, len(code))
assert.True(t, consistOfHexadecimalCharactersOnly(code))

func consistOfHexadecimalCharactersOnly(code string) bool {
	return regexp.MustCompile(`^[0-9a-f]+$`).MatchString(code)
}
*/
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
	u.AssertDeepStackErrorFromRequest(t, err, "email already exists")
}

// TODO !!
func TestCascadingDeletionOfAppsAndVersionsWhenDeletingUser(t *testing.T) {

}

// TODO !!
func TestTolerateTwoUsersWithSamePassword(t *testing.T) {

}

// TODO !! introduce request for user info including username, email, used space -> make a test uploading a version and checking whether the usedSpace field adapted accordingly; same when deleting the version/app, then used space should become zero again
