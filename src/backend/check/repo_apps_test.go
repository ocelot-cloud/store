//go:build unit

package check

import (
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/store"
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"
	"ocelot/store/versions"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	tools.InitializeDatabase()
	code := m.Run()
	os.Exit(code)
}

func TestCreateRepoApp(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.GetAppId(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.True(t, apps.AppRepo.DoesAppExist(appId))
	maintainer, err := apps.AppRepo.GetMaintainerName(appId)
	assert.Nil(t, err)
	assert.Equal(t, tools.SampleUser, maintainer)
}

func TestDeleteAppCascadingThroughUser(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.GetAppId(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.True(t, apps.AppRepo.DoesAppExist(appId))
	assert.Nil(t, apps.AppRepo.DeleteApp(appId))
	assert.False(t, apps.AppRepo.DoesAppExist(appId))
}

func TestDeleteAppDirectly(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.GetAppId(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.True(t, apps.AppRepo.DoesAppExist(appId))
	assert.Nil(t, users.UserRepo.DeleteUser(tools.SampleUser))
	assert.False(t, apps.AppRepo.DoesAppExist(appId))
}

func TestTolerateSameAppsForTwoUsers(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	user2 := tools.SampleUser + "2"
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	newForm := *tools.SampleForm
	newForm.User = user2
	newForm.Email = tools.SampleEmail + "x"
	assert.Nil(t, users.CreateAndValidateUser(&newForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	assert.Nil(t, apps.AppRepo.CreateApp(user2, tools.SampleApp))

	appId1, err := apps.AppRepo.GetAppId(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.True(t, apps.AppRepo.DoesAppExist(appId1))
	appId2, err := apps.AppRepo.GetAppId(user2, tools.SampleApp)
	assert.Nil(t, err)
	assert.True(t, apps.AppRepo.DoesAppExist(appId2))

	assert.Nil(t, apps.AppRepo.DeleteApp(appId1))
	assert.False(t, apps.AppRepo.DoesAppExist(appId1))
	assert.True(t, apps.AppRepo.DoesAppExist(appId2))
}

func TestSearchNegative(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	app := "prefix_myapp_suffix"
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, app))

	appSearchRequest := store.AppSearchRequest{
		SearchTerm:         "some",
		ShowUnofficialApps: true,
	}
	a, err := apps.AppRepo.SearchForApps(appSearchRequest)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(a))
}

func TestSearchingWithEmptySearchTerm(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	emptySearchRequest := store.AppSearchRequest{
		SearchTerm:         "",
		ShowUnofficialApps: true,
	}
	searchedApps, err := apps.AppRepo.SearchForApps(emptySearchRequest)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(searchedApps))

	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.GetAppId(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, nil))

	searchedApps, err = apps.AppRepo.SearchForApps(emptySearchRequest)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(searchedApps))
}

func TestGetAppListRepo(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	list, err := apps.AppRepo.GetAppList(tools.SampleUser)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(list))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp+"x"))
	list, err = apps.AppRepo.GetAppList(tools.SampleUser)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(list))
	assert.Equal(t, tools.SampleApp, list[0].Name)
	assert.Equal(t, tools.SampleApp+"x", list[1].Name)
}

func TestIsAppOwner(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.False(t, apps.AppRepo.IsAppOwner(tools.SampleUser, 1))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.GetAppId(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.True(t, apps.AppRepo.IsAppOwner(tools.SampleUser, appId))

	sampleForm2 := *tools.SampleForm
	sampleForm2.User = tools.SampleUser + "2"
	sampleForm2.Email = tools.SampleEmail + "x"
	assert.Nil(t, users.CreateAndValidateUser(&sampleForm2))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser+"2", tools.SampleApp))
	assert.False(t, apps.AppRepo.IsAppOwner(tools.SampleUser+"2", appId))

	assert.False(t, apps.AppRepo.IsAppOwner("notExistingUser", appId))
}

func TestGetAppName(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.GetAppId(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	name, err := apps.AppRepo.GetAppName(appId)
	assert.Nil(t, err)
	assert.Equal(t, tools.SampleApp, name)
}

func TestCantCreateAppTwiceForSameUser(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	assert.NotNil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
}

func TestCantCreateAppWithoutUser(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.NotNil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
}
