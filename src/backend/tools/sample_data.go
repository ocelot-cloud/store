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

func GetVersionBytesOfSampleUserApp(folderName, sampleUser, sampleApp string, shouldBeValid bool) []byte {
	sampleAppDir := utils.FindDir("assets") + "/" + folderName
	versionBytes, err := validation.ZipDirectory(sampleAppDir)
	if err != nil {
		Logger.Fatal("Failed to read sample version file: %v", err)
	}
	err = validation.ValidateVersion(versionBytes, sampleUser, sampleApp)
	if shouldBeValid && err != nil {
		Logger.Fatal("expected sample version to be valid, but it is not: %v", err)
	}
	if !shouldBeValid && err == nil {
		Logger.Fatal("expected sample version to be invalid, but it is valid")
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
