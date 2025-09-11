//go:build component

package check

import (
	"ocelot/store/tools"
	"ocelot/store/versions"
	"testing"

	"github.com/ocelot-cloud/shared/assert"
	u "github.com/ocelot-cloud/shared/utils"
)

func TestVersionDownload(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail))
	assert.Nil(t, hub.Login(tools.SampleUser, tools.SamplePassword))

	notExistingVersionId := "0"
	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	_, err = hub.DownloadVersion(notExistingVersionId)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, versions.VersionDoesNotExistError)

	versionId, err := hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)
	foundVersions, err := hub.GetVersions(appId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundVersions))
	assert.Equal(t, tools.SampleVersion, foundVersions[0].Name)

	fullVersionInfo, err := hub.DownloadVersion(versionId)
	assert.Nil(t, err)
	assert.Equal(t, SampleVersionFileContent, fullVersionInfo.Content)
}

func TestUploadVersion(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	notExistingVersionId := "0"
	_, err := hub.UploadVersion(notExistingVersionId, tools.SampleVersion, SampleVersionFileContent)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "app does not exist")

	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	versionId, err := hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)

	_, err = hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "version already exists")

	versions, err := hub.GetVersions(appId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(versions))
	assert.Equal(t, tools.SampleVersion, versions[0].Name)

	assert.Nil(t, hub.DeleteVersion(versionId))
	versions, err = hub.GetVersions(appId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(versions))

	err = hub.DeleteVersion(versionId)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, "version does not exist")
}

// TODO !! not sure what this test is asserting? why unhappy path?
func TestGetVersionsUnhappyPath(t *testing.T) {
	hub := GetHub()
	defer hub.WipeData()
	assert.Nil(t, hub.RegisterAndValidateUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail))
	assert.Nil(t, hub.Login(tools.SampleUser, tools.SamplePassword))

	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	versionList, err := hub.GetVersions(appId)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(versionList))
}

// TODO !! two download versions, is that duplication?
func TestDownloadDummyVersion(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()
	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	versionId, err := hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)
	apps, err := hub.SearchForApps(tools.SampleApp, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	versions, err := hub.GetVersions(appId)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(versions))

	info, err := hub.DownloadVersion(versionId)
	assert.Nil(t, err)
	assert.Equal(t, tools.SampleUser, info.Maintainer)
	assert.Equal(t, tools.SampleApp, info.AppName)
	assert.Equal(t, tools.SampleVersion, info.VersionName)
	assert.True(t, len(info.Content) > 100)
}
