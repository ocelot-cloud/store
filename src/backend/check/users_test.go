//go:build component

package check

import (
	"ocelot/store/tools"
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
	u.AssertDeepStackErrorFromRequest(t, err, "incorrect username or password")

	hub.Parent.Cookie = nil
	err = hub.Login(tools.SampleUser, newPassword)
	assert.Nil(t, err)
	assert.NotNil(t, hub.Parent.Cookie)
}

// TODO !! better name: test user cannot be registered twice
func TestRegistration(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail))
	err := hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "user already exists")
}

func TestLogout(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	_, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	assert.Nil(t, hub.Logout())
	_, err = hub.CreateApp(tools.SampleApp)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "cookie not found")
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
	u.AssertDeepStackErrorFromRequest(t, err, "email already exists")
}

// TODO !!
func TestCascadingDeletionOfAppsAndVersionsWhenDeletingUser(t *testing.T) {

}
