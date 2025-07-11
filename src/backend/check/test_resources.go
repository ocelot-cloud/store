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
	assert.Nil(t, hub.RegisterUser(tools.SampleUser, tools.SamplePassword, tools.SampleEmail))
	assert.Nil(t, hub.ValidateCode())
	err := hub.Login(tools.SampleUser, tools.SamplePassword)
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
			SetCookieHeader: true,
			RootUrl:         tools.RootUrl,
		},
	}
}
