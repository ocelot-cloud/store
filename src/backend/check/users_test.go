//go:build component

package check

import (
	"ocelot/store/tools"
	"testing"

	"github.com/ocelot-cloud/shared/assert"
)

func TestLogin(t *testing.T) {
	hub := GetHub()
	err := hub.Login(tools.SampleUser, tools.SamplePassword)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "user does not exist", 400)
}

func TestChangePassword(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	newPassword := tools.SamplePassword + "x"

	assert.Nil(t, hub.ChangePassword(tools.SamplePassword, newPassword))
	err := hub.Login(tools.SampleUser, tools.SamplePassword)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "incorrect username or password", 400)

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
	AssertDeepStackErrorWithCode(t, err, "user already exists", 400)
}

func TestLogout(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	_, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	assert.Nil(t, hub.Logout())
	_, err = hub.CreateApp(tools.SampleApp)
	assert.NotNil(t, err)
	AssertDeepStackErrorWithCode(t, err, "cookie not found", 400)
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
	AssertDeepStackErrorWithCode(t, err, "email already exists", 400)
}
