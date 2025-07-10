//go:build unit

package check

import (
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/utils"
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"
	"ocelot/store/versions"
	"regexp"
	"testing"
	"time"
)

func TestCreateRepoUser(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.False(t, users.UserRepo.DoesUserExist(tools.SampleUser))
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.True(t, users.UserRepo.DoesUserExist(tools.SampleUser))

	assert.Nil(t, users.UserRepo.DeleteUser(tools.SampleUser))
	assert.False(t, users.UserRepo.DoesUserExist(tools.SampleUser))
}

func TestCantCreateUserTwice(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.NotNil(t, users.CreateAndValidateUser(tools.SampleForm))
}

func TestTolerateSamePasswordForTwoUsers(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	user2 := tools.SampleUser + "2"
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	newForm := *tools.SampleForm
	newForm.User = user2
	newForm.Email = tools.SampleEmail + "x"
	assert.Nil(t, users.CreateAndValidateUser(&newForm))
	assert.True(t, users.UserRepo.IsPasswordCorrect(tools.SampleUser, tools.SamplePassword))
	assert.True(t, users.UserRepo.IsPasswordCorrect(user2, tools.SamplePassword))
}

func TestPasswordVerification(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.True(t, users.UserRepo.IsPasswordCorrect(tools.SampleUser, tools.SamplePassword))
	assert.False(t, users.UserRepo.IsPasswordCorrect(tools.SampleUser, tools.SamplePassword+"x"))
}

func TestCookieExpiration(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	_, err := users.UserRepo.GetUserViaCookie("")
	assert.NotNil(t, err)

	assert.True(t, users.UserRepo.IsCookieExpired("non-existing-cookie"))

	timeIn30Days := utils.GetTimeInSevenDays()
	cookie, _ := utils.GenerateCookie()
	assert.Nil(t, users.UserRepo.HashAndSaveCookie(tools.SampleUser, cookie.Value, timeIn30Days))
	assert.False(t, users.UserRepo.IsCookieExpired(cookie.Value))

	past := time.Now().Add(-1 * time.Second)
	assert.Nil(t, users.UserRepo.HashAndSaveCookie(tools.SampleUser, cookie.Value, past))
	assert.True(t, users.UserRepo.IsCookieExpired(cookie.Value))

	user, err := users.UserRepo.GetUserViaCookie(cookie.Value)
	assert.Nil(t, err)
	assert.Equal(t, tools.SampleUser, user)
}

func TestChangeRepoPassword(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.True(t, users.UserRepo.IsPasswordCorrect(tools.SampleUser, tools.SamplePassword))
	newPassword := tools.SamplePassword + "x"
	assert.Nil(t, users.UserRepo.ChangePassword(tools.SampleUser, newPassword))
	assert.False(t, users.UserRepo.IsPasswordCorrect(tools.SampleUser, tools.SampleForm.Password))
	assert.True(t, users.UserRepo.IsPasswordCorrect(tools.SampleUser, newPassword))
}

func TestRepoLogout(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	sampleCookie := "asdasdasd"
	err := users.UserRepo.HashAndSaveCookie(tools.SampleUser, sampleCookie, time.Now().Add(1*time.Hour))
	assert.Nil(t, err)
	assert.False(t, users.UserRepo.IsCookieExpired(sampleCookie))
	assert.Nil(t, users.UserRepo.Logout(tools.SampleUser))
	assert.True(t, users.UserRepo.IsCookieExpired(sampleCookie))
}

func TestEmailDuringUserCreation(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.False(t, users.UserRepo.DoesEmailExist(tools.SampleEmail))
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.True(t, users.UserRepo.DoesEmailExist(tools.SampleEmail))

	newForm := *tools.SampleForm
	newForm.User = tools.SampleUser + "x"
	code, err := users.UserRepo.CreateUser(&newForm)
	assert.Nil(t, err)
	assert.NotNil(t, users.UserRepo.ValidateUser(code))
	assert.False(t, users.UserRepo.DoesUserExist(tools.SampleUser+"x"))

	newForm.Email = tools.SampleEmail + "x"
	code, err = users.UserRepo.CreateUser(&newForm)
	assert.Nil(t, err)
	assert.Nil(t, users.UserRepo.ValidateUser(code))
	assert.True(t, users.UserRepo.DoesUserExist(tools.SampleUser+"x"))
}

func TestValidationCode(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	code, err := users.UserRepo.CreateUser(tools.SampleForm)
	assert.Nil(t, err)
	assert.Equal(t, 64, len(code))
	assert.True(t, consistOfHexadecimalCharactersOnly(code))

	assert.False(t, users.UserRepo.DoesUserExist(tools.SampleUser))
	err = users.UserRepo.ValidateUser(code)
	assert.Nil(t, err)
	assert.True(t, users.UserRepo.DoesUserExist(tools.SampleUser))

	err = users.UserRepo.ValidateUser(code)
	assert.NotNil(t, err)
	assert.Equal(t, "code not found", err.Error())
}

func consistOfHexadecimalCharactersOnly(code string) bool {
	return regexp.MustCompile(`^[0-9a-f]+$`).MatchString(code)
}

func TestSpace(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))

	tenMegaBytes := 10 * 1024 * 1024
	assert.Nil(t, users.UserRepo.IsThereEnoughSpaceToAddVersion(tools.SampleUser, tenMegaBytes))
	assert.NotNil(t, users.UserRepo.IsThereEnoughSpaceToAddVersion(tools.SampleUser, tenMegaBytes+1))

	appId, err := apps.AppRepo.GetAppId(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	oneKiloByte := 1024
	randomBytes := make([]byte, oneKiloByte)
	assert.Nil(t, err)
	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, "version", randomBytes))

	assert.Nil(t, users.UserRepo.IsThereEnoughSpaceToAddVersion(tools.SampleUser, tenMegaBytes-oneKiloByte))
	assert.NotNil(t, users.UserRepo.IsThereEnoughSpaceToAddVersion(tools.SampleUser, tenMegaBytes-oneKiloByte+1))
}

func TestUsedSpace(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	space, err := users.UserRepo.GetUsedSpaceInBytes(tools.SampleUser)
	assert.Nil(t, err)
	assert.Equal(t, 0, space)

	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.GetAppId(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)

	bytes := []byte("hello")
	bytes2 := []byte(" world")
	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, bytes))
	space, err = users.UserRepo.GetUsedSpaceInBytes(tools.SampleUser)
	assert.Nil(t, err)
	assert.Equal(t, 5, space)

	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion+"x", bytes2))
	space, err = users.UserRepo.GetUsedSpaceInBytes(tools.SampleUser)
	assert.Nil(t, err)
	assert.Equal(t, 11, space)

	versionId, err := versions.VersionRepo.GetVersionId(appId, tools.SampleVersion)
	assert.Nil(t, err)
	assert.Nil(t, versions.VersionRepo.DeleteVersion(versionId))
	space, err = users.UserRepo.GetUsedSpaceInBytes(tools.SampleUser)
	assert.Nil(t, err)
	assert.Equal(t, 6, space)

	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, bytes2))
	space, err = users.UserRepo.GetUsedSpaceInBytes(tools.SampleUser)
	assert.Nil(t, err)
	assert.Equal(t, 12, space)

	assert.Nil(t, apps.AppRepo.DeleteApp(appId))
	space, err = users.UserRepo.GetUsedSpaceInBytes(tools.SampleUser)
	assert.Nil(t, err)
	assert.Equal(t, 0, space)
}
