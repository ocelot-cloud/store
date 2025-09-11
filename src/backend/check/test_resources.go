package check

import (
	"ocelot/store/tools"
	"ocelot/store/users"
	"os"
	"testing"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
)

var SampleVersionFileContent = GetValidVersionBytesOfSampleMaintainerApp()

func GetValidVersionBytesOfSampleMaintainerApp() []byte {
	assetsDir, err := u.FindDir("assets")
	if err != nil {
		u.Logger.Error("Failed to find assets directory", deepstack.ErrorField, err)
		// TODO !! return error
	}
	sampleAppDir := assetsDir + "/sample-app"
	versionBytes, err := validation.ZipDirectory(sampleAppDir)
	if err != nil {
		u.Logger.Error("Failed to read sample version file", deepstack.ErrorField, err)
		os.Exit(1)
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
	hub.WipeData()
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
