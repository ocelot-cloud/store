package check

import (
	"ocelot/store/tools"
	"ocelot/store/users"
	"testing"

	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
)

var SampleVersionFileContent = GetValidVersionBytesOfSampleMaintainerApp()

func GetValidVersionBytesOfSampleMaintainerApp() []byte {
	assetsDir, err := u.FindDir("assets")
	if err != nil {
		panic("Failed to find assets directory")
	}
	sampleAppDir := assetsDir + "/sampleapp"
	versionBytes, err := validation.ZipDirectory(sampleAppDir)
	if err != nil {
		panic("Failed to read sample version file")
	}
	return versionBytes
}

func GetHubAndLogin(t *testing.T) *store.AppStoreClientImpl {
	client := GetHub()
	assert.Nil(t, client.RegisterUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail))
	assert.Nil(t, client.ValidateCode())
	err := client.Login(tools.SampleUser, tools.SamplePassword)
	assert.Nil(t, err)
	return client
}

func GetHub() *store.AppStoreClientImpl {
	hub := createHubClient()
	return hub
}

func createHubClient() *store.AppStoreClientImpl {
	return &store.AppStoreClientImpl{
		Parent: u.ComponentClient{
			SetCookieHeader: true,
			RootUrl:         users.DefaultAppStoreHost,
		},
	}
}
