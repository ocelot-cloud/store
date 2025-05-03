package tools

import (
	"github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
)

var (
	SampleUser     = "samplemaintainer"
	SampleApp      = "gitea"
	SampleVersion  = "0.0.1"
	SampleEmail    = "testuser@example.com"
	SamplePassword = "mypassword"
	SampleForm     = &RegistrationForm{
		SampleUser,
		SamplePassword,
		SampleEmail,
	}
)

func GetValidVersionBytes() []byte {
	sampleAppDir := utils.FindDir("assets") + "/sample_app"
	versionBytes, err := validation.ZipDirectory(sampleAppDir)
	if err != nil {
		Logger.Fatal("Failed to read sample version file: %v", err)
	}
	return versionBytes
}
