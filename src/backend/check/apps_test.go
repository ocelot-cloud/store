//go:build component

package check

import (
	"ocelot/store/apps"
	"ocelot/store/tools"
	"testing"

	"github.com/ocelot-cloud/shared/assert"
	u "github.com/ocelot-cloud/shared/utils"
)

func TestCreateAndDeleteApp(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	foundApps, err := hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundApps))
	foundApp := foundApps[0]
	assert.Equal(t, tools.SampleApp, foundApp.Name)

	_, err = hub.CreateApp(tools.SampleApp)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps.AppAlreadyExistsError)

	assert.Nil(t, hub.DeleteApp(appId))
	foundApps, err = hub.ListOwnApps()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundApps))
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

func TestCreationOfOcelotCloudAppIsForbidden(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	_, err := hub.CreateApp("ocelotcloud")
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, apps.AppNameReservedError)
}

// TODO !! all search cases covered?
// TODO !! make sure that new search method is tested sufficiently; search via maintainer AND app separately, and one test should cover both at the same time

func TestUnofficialAppFilteringWhenSearching(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	_, err = hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)

	apps, err := hub.SearchForApps("", tools.SampleApp, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, appId, apps[0].AppId)

	apps, err = hub.SearchForApps("", tools.SampleApp, false)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(apps))
}

func TestAllowEmptyStringAsSearchTerm(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	apps, err := hub.SearchForApps("", "", true)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(apps))

	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	_, err = hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)

	apps, err = hub.SearchForApps("", "", true)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
}

// TODO !!
func TestCascadingDeletionOfVersionsWhenDeletingApp(t *testing.T) {

}

// TODO
func TestTolerateTwoDifferentUsersCreateAppWithSameName(t *testing.T) {

}

// TODO !!
func TestSearchForNonExistingAppsReturnsEmptyList(t *testing.T) {

}

// TODO !!
func TestCantCreateAppTwiceForSameUser(t *testing.T) {

}

// TODO !! when searching apps and there are two versions, assert that the latest is returned
