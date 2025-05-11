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

func GetValidVersionBytesOfSampleUserApp(sampleUser, sampleApp string) []byte {
	sampleAppDir := utils.FindDir("assets") + "/sampleuser-app"
	versionBytes, err := validation.ZipDirectory(sampleAppDir)
	if err != nil {
		Logger.Fatal("Failed to read sample version file: %v", err)
	}
	err = validation.ValidateVersion(versionBytes, sampleUser, sampleApp)
	if err != nil {
		Logger.Fatal("Failed to validate sample version file: %v", err)
		return nil
	}
	return versionBytes
}

func GetValidVersionBytesOfSampleMaintainerApp() []byte {
	sampleAppDir := utils.FindDir("assets") + "/samplemaintainer-app"
	versionBytes, err := validation.ZipDirectory(sampleAppDir)
	if err != nil {
		Logger.Fatal("Failed to read sample version file: %v", err)
	}
	return versionBytes
}
