package check

import (
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/store"
	"github.com/ocelot-cloud/shared/utils"
	"ocelot/store/tools"
	"testing"
)

func GetHubAndLogin(t *testing.T) *store.AppStoreClient {
	hub := GetHub()
	assert.Nil(t, hub.RegisterUser())
	assert.Nil(t, hub.ValidateCode())
	err := hub.Login()
	assert.Nil(t, err)
	return hub
}

func GetHub() *store.AppStoreClient {
	hub := createHubClient()
	hub.WipeData()
	return hub
}

var SampleVersionFileContent = tools.GetValidVersionBytesOfSampleMaintainerApp()

func createHubClient() *store.AppStoreClient {
	return &store.AppStoreClient{
		Parent: utils.ComponentClient{
			User:            tools.SampleUser,
			Password:        tools.SamplePassword,
			SetCookieHeader: true,
			RootUrl:         tools.RootUrl,
		},

		Email:              tools.SampleEmail,
		App:                tools.SampleApp,
		Version:            tools.SampleVersion,
		UploadContent:      SampleVersionFileContent,
		AppId:              "0",
		VersionId:          "0",
		ValidationCode:     "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		ShowUnofficialApps: true,
	}
}
