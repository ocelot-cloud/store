package tools

import (
	"os"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	"github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
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

// TODO !! return error
func GetVersionBytesOfSampleUserApp(folderName, sampleUser, sampleApp string, shouldBeValid bool) []byte {
	assetsDir, err := utils.FindDir("assets") // TODO !! the DirectoryProvider Service should find the assets directory once at the beginning
	if err != nil {
		Logger.Error("Failed to find assets directory", deepstack.ErrorField, err)
		os.Exit(1)
	}
	sampleAppDir := assetsDir + "/" + folderName
	versionBytes, err := validation.ZipDirectory(sampleAppDir)
	if err != nil {
		Logger.Error("Failed to read sample version file", deepstack.ErrorField, err)
		os.Exit(1)
	}
	err = validation.ValidateVersion(versionBytes, sampleUser, sampleApp)
	if shouldBeValid && err != nil {
		Logger.Error("expected sample version to be valid, but it is not", deepstack.ErrorField, err)
		os.Exit(1)
	}
	if !shouldBeValid && err == nil {
		Logger.Error("expected sample version to be invalid, but it is valid")
		os.Exit(1)
	}
	return versionBytes
}

func GetValidVersionBytesOfSampleMaintainerApp() []byte {
	assetsDir, err := utils.FindDir("assets")
	if err != nil {
		Logger.Error("Failed to find assets directory", deepstack.ErrorField, err)
		// TODO !! return error
	}
	// TODO !! simple call it "sampleapp"
	sampleAppDir := assetsDir + "/samplemaintainer-app"
	versionBytes, err := validation.ZipDirectory(sampleAppDir)
	if err != nil {
		Logger.Error("Failed to read sample version file", deepstack.ErrorField, err)
		os.Exit(1)
	}
	return versionBytes
}
