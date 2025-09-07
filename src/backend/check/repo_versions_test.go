//go:build unit

package check

/* TODO !! move to unit and component tests
import (
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/store"
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"
	"ocelot/store/versions"
	"sort"
	"strconv"
	"testing"
	"time"
)

func TestCreateRepoVersion(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.GetAppId2(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, []byte("asdf")))
	versionId, err := versions.VersionRepo.GetVersionId(appId, tools.SampleVersion)
	assert.Nil(t, err)
	assert.True(t, versions.VersionRepo.DoesVersionExist(versionId))
	versions, err := versions.VersionRepo.GetVersionList(appId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(versions))
	version := versions[0]
	assert.Equal(t, strconv.Itoa(versionId), version.Id)
	assert.Equal(t, tools.SampleVersion, version.Name)
	assert.True(t, version.CreationTimestamp.Before(time.Now().UTC()))
	assert.True(t, version.CreationTimestamp.After(time.Now().UTC().Add(-1*time.Second)))
}

func TestGetVersionList(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.GetAppId2(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	foundVersions, err := versions.VersionRepo.GetVersionList(appId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundVersions))
	versionId, err := versions.VersionRepo.GetVersionId(appId, tools.SampleVersion)
	assert.NotNil(t, err)
	assert.False(t, versions.VersionRepo.DoesVersionExist(versionId))

	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, []byte("asdf")))
	foundVersions, err = versions.VersionRepo.GetVersionList(appId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundVersions))
	assert.Equal(t, tools.SampleVersion, foundVersions[0].Name)
	versionId, err = versions.VersionRepo.GetVersionId(appId, tools.SampleVersion)
	assert.Nil(t, err)
	assert.True(t, versions.VersionRepo.DoesVersionExist(versionId))
	data, err := versions.VersionRepo.GetVersionContent(versionId)
	assert.Nil(t, err)
	assert.Equal(t, []byte("asdf"), data)
	assert.True(t, foundVersions[0].CreationTimestamp.Before(time.Now().UTC()))
	assert.True(t, foundVersions[0].CreationTimestamp.After(time.Now().UTC().Add(-1*time.Second)))

	assert.Nil(t, versions.VersionRepo.DeleteVersion(versionId))
	foundVersions, err = versions.VersionRepo.GetVersionList(appId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundVersions))
	assert.False(t, versions.VersionRepo.DoesVersionExist(versionId))

	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, []byte("asdf")))
	foundVersions, err = versions.VersionRepo.GetVersionList(appId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundVersions))
	assert.Equal(t, tools.SampleVersion, foundVersions[0].Name)
	versionId, err = versions.VersionRepo.GetVersionId(appId, tools.SampleVersion)
	assert.Nil(t, err)
	assert.True(t, versions.VersionRepo.DoesVersionExist(versionId))
}

func TestGetVersionListForNonExistingVersions(t *testing.T) {
	list, err := versions.VersionRepo.GetVersionList(-1)
	assert.NotNil(t, err)
	assert.Nil(t, list)
}

func TestAppIdConsistency(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.GetAppId2(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, []byte("asdf")))
	versionId, err := versions.VersionRepo.GetVersionId(appId, tools.SampleVersion)
	assert.Nil(t, err)

	appList, err := apps.AppRepo.GetAppList(tools.SampleUser)
	assert.Nil(t, err)
	assert.Equal(t, strconv.Itoa(appId), appList[0].Id)

	versionList, err := versions.VersionRepo.GetVersionList(appId)
	assert.Nil(t, err)
	assert.Equal(t, strconv.Itoa(versionId), versionList[0].Id)
}

func TestIsVersionOwner(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.False(t, versions.VersionRepo.IsVersionOwner(tools.SampleUser, 1))

	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.GetAppId2(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.False(t, versions.VersionRepo.IsVersionOwner(tools.SampleUser, 1))

	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, []byte("asdf")))
	versionId, err := versions.VersionRepo.GetVersionId(appId, tools.SampleVersion)
	assert.Nil(t, err)
	assert.True(t, versions.VersionRepo.IsVersionOwner(tools.SampleUser, versionId))

	sampleForm2 := *tools.SampleForm
	sampleForm2.User = tools.SampleUser + "2"
	sampleForm2.Email = tools.SampleEmail + "2"
	assert.Nil(t, users.CreateAndValidateUser(&sampleForm2))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser+"2", tools.SampleApp))
	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, []byte("asdf")))
	assert.False(t, versions.VersionRepo.IsVersionOwner(tools.SampleUser+"2", appId))

	assert.False(t, versions.VersionRepo.IsVersionOwner("notExistingUser", versionId))
}

func TestGetAppIdByVersionId(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	expectedAppId, err := apps.AppRepo.GetAppId2(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.Nil(t, versions.VersionRepo.CreateVersion(expectedAppId, tools.SampleVersion, []byte("asdf")))
	versionId, err := versions.VersionRepo.GetVersionId(expectedAppId, tools.SampleVersion)
	assert.Nil(t, err)

	actualAppId, err := versions.VersionRepo.GetAppIdByVersionId(versionId)
	assert.Nil(t, err)
	assert.Equal(t, expectedAppId, actualAppId)
}

func TestGetFullVersionInfo(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.GetAppId2(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, SampleVersionFileContent))
	versionId, err := versions.VersionRepo.GetVersionId(appId, tools.SampleVersion)
	assert.Nil(t, err)

	fullVersionInfo, err := versions.VersionRepo.GetFullVersionInfo(versionId)
	assert.Nil(t, err)
	assert.Equal(t, tools.SampleUser, fullVersionInfo.Maintainer)
	assert.Equal(t, tools.SampleApp, fullVersionInfo.AppName)
	assert.Equal(t, tools.SampleVersion, fullVersionInfo.VersionName)
	assert.Equal(t, SampleVersionFileContent, fullVersionInfo.Content)
	assert.Equal(t, versionId, fullVersionInfo.Id)
	assert.True(t, time.Now().UTC().Add(-1*time.Minute).Before(fullVersionInfo.VersionCreationTimestamp))
	assert.True(t, time.Now().UTC().Add(1*time.Minute).After(fullVersionInfo.VersionCreationTimestamp))
}

func TestSearchForApps(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	app1 := "prefix_myapp_suffix"
	app2 := "prefix_another-app_suffix"
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, app1))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, app2))

	app1Id, err := apps.AppRepo.GetAppId2(tools.SampleUser, app1)
	assert.Nil(t, err)
	err = versions.VersionRepo.CreateVersion(app1Id, tools.SampleVersion, []byte("asdf"))
	assert.Nil(t, err)
	app2Id, err := apps.AppRepo.GetAppId2(tools.SampleUser, app2)
	assert.Nil(t, err)
	sampleVersion2 := tools.SampleVersion + "x"
	err = versions.VersionRepo.CreateVersion(app2Id, sampleVersion2, []byte("asdf"))
	assert.Nil(t, err)

	appSearchRequest := store.AppSearchRequest{
		SearchTerm:         "app",
		ShowUnofficialApps: true,
	}
	foundApps, err := apps.AppRepo.SearchForApps(appSearchRequest)
	assert.Nil(t, err)
	sort.Slice(foundApps, func(i, j int) bool {
		return foundApps[i].AppName < foundApps[j].AppName
	})

	assert.Equal(t, 2, len(foundApps))
	assert.Equal(t, tools.SampleUser, foundApps[0].Maintainer)
	assert.Equal(t, tools.SampleUser, foundApps[1].Maintainer)
	assert.Equal(t, app2, foundApps[0].AppName)
	assert.Equal(t, app1, foundApps[1].AppName)
}

func TestSearchForApps_LatestVersions(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appSearchRequest := store.AppSearchRequest{
		SearchTerm:         tools.SampleApp,
		ShowUnofficialApps: true,
	}
	searchedApps, err := apps.AppRepo.SearchForApps(appSearchRequest)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(searchedApps))

	appId, err := apps.AppRepo.GetAppId2(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, []byte("asdf")))
	versionId, err := versions.VersionRepo.GetVersionId(appId, tools.SampleVersion)
	assert.Nil(t, err)
	searchedApps, err = apps.AppRepo.SearchForApps(appSearchRequest)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(searchedApps))
	assert.Equal(t, strconv.Itoa(appId), searchedApps[0].AppId)
	assert.Equal(t, strconv.Itoa(versionId), searchedApps[0].LatestVersionId)
	assert.Equal(t, tools.SampleVersion, searchedApps[0].LatestVersionName)

	sampleVersion2 := tools.SampleVersion + "x"
	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, sampleVersion2, []byte("asdf")))
	version2Id, err := versions.VersionRepo.GetVersionId(appId, sampleVersion2)
	assert.Nil(t, err)
	searchedApps, err = apps.AppRepo.SearchForApps(appSearchRequest)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(searchedApps))
	assert.Equal(t, strconv.Itoa(appId), searchedApps[0].AppId)
	assert.Equal(t, strconv.Itoa(version2Id), searchedApps[0].LatestVersionId)
	assert.Equal(t, sampleVersion2, searchedApps[0].LatestVersionName)
}

func TestUnofficialAppFiltering(t *testing.T) {
	defer users.UserRepo.WipeDatabase()
	officialUser := "ocelotcloud"
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	officialUserRegistrationForm := &store.RegistrationForm{
		User:     officialUser,
		Password: "password",
		Email:    officialUser + "@ocelot-cloud.org",
	}
	assert.Nil(t, users.CreateAndValidateUser(officialUserRegistrationForm))

	appSearchRequest := store.AppSearchRequest{
		SearchTerm:         "app",
		ShowUnofficialApps: true,
	}
	foundApps, err := apps.AppRepo.SearchForApps(appSearchRequest)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundApps))

	app1 := "official_app"
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, app1))
	app1Id, err := apps.AppRepo.GetAppId2(tools.SampleUser, app1)
	assert.Nil(t, err)
	assert.Nil(t, versions.VersionRepo.CreateVersion(app1Id, tools.SampleVersion, []byte("sample-bytes")))

	app2 := "unofficial_app"
	assert.Nil(t, apps.AppRepo.CreateApp(officialUser, app2))
	app2Id, err := apps.AppRepo.GetAppId2(officialUser, app2)
	assert.Nil(t, err)
	assert.Nil(t, versions.VersionRepo.CreateVersion(app2Id, tools.SampleVersion, []byte("sample-bytes")))

	appSearchRequest = store.AppSearchRequest{
		SearchTerm:         "app",
		ShowUnofficialApps: true,
	}
	foundApps, err = apps.AppRepo.SearchForApps(appSearchRequest)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundApps))

	appSearchRequest = store.AppSearchRequest{
		SearchTerm:         "app",
		ShowUnofficialApps: false,
	}
	foundApps, err = apps.AppRepo.SearchForApps(appSearchRequest)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundApps))
	assert.Equal(t, officialUser, foundApps[0].Maintainer)
}


*/
