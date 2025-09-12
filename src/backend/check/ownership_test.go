//go:build component

package check

import (
	"testing"
)

// TODO !! check whether other file already test ownership, and centralize it here
// TODO !! test cases from this: you can not operate on apps or versions you dont own -> deletion/creation(or upload) etc. make research which operations are affected

// TODO !! use case: delete app you do not own; delete version you dont own; upload version to app you dont own

func TestDeleteAppYouDoNotOwn(t *testing.T) {
	/* TODO !!
	client1 := GetHub()
	defer client1.WipeData()
	err := client1.RegisterAndValidateUser("user1", "password", "user1@sample.com")
	assert.Nil(t, err)
	assert.Nil(t, client1.Login("user1", "password"))

	client2 := GetHub()
	err = client2.RegisterAndValidateUser("user2", "password", "user2@sample.com")
	assert.Nil(t, err)
	assert.Nil(t, client2.Login("user2", "password"))

	appId, err := client1.CreateApp("sampleapp")
	assert.Nil(t, err)
	err = client2.DeleteApp(appId)
	u.AssertDeepStackErrorFromRequest(t, err, apps.YouDoNotOwnThisAppError)
	*/
}

/* material

func TestIsAppOwner(t *testing.T) {
	defer users.UserRepo.WipeUsers()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.False(t, apps.AppRepo.DoesUserOwnApp(tools.SampleUser, 1))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.DoesAppExist(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.True(t, apps.AppRepo.DoesUserOwnApp(tools.SampleUser, appId))

	sampleForm2 := *tools.SampleForm
	sampleForm2.User = tools.SampleUser + "2"
	sampleForm2.Email = tools.SampleEmail + "x"
	assert.Nil(t, users.CreateAndValidateUser(&sampleForm2))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser+"2", tools.SampleApp))
	assert.False(t, apps.AppRepo.DoesUserOwnApp(tools.SampleUser+"2", appId))

	assert.False(t, apps.AppRepo.DoesUserOwnApp("notExistingUser", appId))
}


func TestOwnership(t *testing.T) {
	hub := GetHub()
	testVersionOwnership(t, hub, hub.DeleteApp)
	hub = GetHub()
	testVersionOwnership(t, hub, hub.UploadVersion)
}

func testVersionOwnership(t *testing.T, hub *store.AppStoreClient, operation func() error) {
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
	assert.Equal(t, u.GetErrMsg(400, apps.YouDoNotOwnThisAppError), err.Error())
}

func TestOwnershipOfDeleteVersion(t *testing.T) {
	hub := GetHubAndLogin(t)
	defer hub.WipeData()

	// TODO this block occurs quite often, can be abstracted
	appId, err := hub.CreateApp(tools.SampleApp)
	assert.Nil(t, err)
	versionId, err := hub.UploadVersion(appId, tools.SampleVersion, SampleVersionFileContent)
	assert.Nil(t, err)

	assert.Nil(t, hub.RegisterAndValidateUser(tools.SampleUser+"2", tools.SamplePassword, tools.SampleEmail+"x"))
	assert.Nil(t, hub.Login(tools.SampleUser+"2", tools.SamplePassword))

	err = hub.DeleteVersion(versionId)
	assert.NotNil(t, err)
	u.AssertDeepStackErrorFromRequest(t, err, versions.NotOwningThisVersionError)
}

func TestIsVersionOwner(t *testing.T) {
	defer users.UserRepo.WipeUsers()
	assert.Nil(t, users.CreateAndValidateUser(tools.SampleForm))
	assert.False(t, versions.VersionRepo.DoesUserOwnVersion(tools.SampleUser, 1))

	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser, tools.SampleApp))
	appId, err := apps.AppRepo.DoesAppExist(tools.SampleUser, tools.SampleApp)
	assert.Nil(t, err)
	assert.False(t, versions.VersionRepo.DoesUserOwnVersion(tools.SampleUser, 1))

	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, []byte("asdf")))
	versionId, err := versions.VersionRepo.DoesVersionNameExist(appId, tools.SampleVersion)
	assert.Nil(t, err)
	assert.True(t, versions.VersionRepo.DoesUserOwnVersion(tools.SampleUser, versionId))

	sampleForm2 := *tools.SampleForm
	sampleForm2.User = tools.SampleUser + "2"
	sampleForm2.Email = tools.SampleEmail + "2"
	assert.Nil(t, users.CreateAndValidateUser(&sampleForm2))
	assert.Nil(t, apps.AppRepo.CreateApp(tools.SampleUser+"2", tools.SampleApp))
	assert.Nil(t, versions.VersionRepo.CreateVersion(appId, tools.SampleVersion, []byte("asdf")))
	assert.False(t, versions.VersionRepo.DoesUserOwnVersion(tools.SampleUser+"2", appId))

	assert.False(t, versions.VersionRepo.DoesUserOwnVersion("notExistingUser", versionId))
}
*/
