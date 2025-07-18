package tools

import (
	"github.com/ocelot-cloud/shared/store"
	"github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
	"os"
)

var (
	SampleUser     = "samplemaintainer"
	SampleApp      = "gitea"
	SampleVersion  = "0.0.1"
	SampleEmail    = "testuser@example.com"
	SamplePassword = "mypassword"
	SampleForm     = &store.RegistrationForm{
		SampleUser,
		SamplePassword,
		SampleEmail,
	}
)

func GetVersionBytesOfSampleUserApp(folderName, sampleUser, sampleApp string, shouldBeValid bool) []byte {
	sampleAppDir := utils.FindDir("assets") + "/" + folderName
	versionBytes, err := validation.ZipDirectory(sampleAppDir)
	if err != nil {
		Logger.Error("Failed to read sample version file", utils.ErrorField, err)
		os.Exit(1)
	}
	err = validation.ValidateVersion(versionBytes, sampleUser, sampleApp)
	if shouldBeValid && err != nil {
		Logger.Error("expected sample version to be valid, but it is not", utils.ErrorField, err)
		os.Exit(1)
	}
	if !shouldBeValid && err == nil {
		Logger.Error("expected sample version to be invalid, but it is valid")
		os.Exit(1)
	}
	return versionBytes
}

func GetValidVersionBytesOfSampleMaintainerApp() []byte {
	sampleAppDir := utils.FindDir("assets") + "/samplemaintainer-app"
	versionBytes, err := validation.ZipDirectory(sampleAppDir)
	if err != nil {
		Logger.Error("Failed to read sample version file", utils.ErrorField, err)
		os.Exit(1)
	}
	return versionBytes
}
