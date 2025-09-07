package main

import (
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"
	"ocelot/store/versions"
	"os"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
)

type DatabaseSampleDataSeeder struct {
	AppRepo     apps.AppRepository
	VersionRepo versions.VersionRepository
	UserRepo    users.UserRepository
}

func (d *DatabaseSampleDataSeeder) SeedSampleDataForTestMode() {
	// TODO !! should be called in main
	if tools.Profile == tools.TEST {
		d.UserRepo.WipeDatabase()
		// This user is created to manually test the GUI so that account registration can be skipped to save time.
		sampleUser := "sample"
		// The user may already exist from previous runs. In this case, ignore the error.
		err := d.UserRepo.CreateAndValidateUser(&store.RegistrationForm{
			User:     sampleUser,
			Password: "password",
			Email:    "sample@sample.com",
		})
		if err != nil {
			u.Logger.Debug("Failed to create user - maybe because he already exists, error", tools.UserField, sampleUser, deepstack.ErrorField, err)
		}
		u.Logger.Warn("created user with weak password for manual testing", tools.UserField, sampleUser)
		d.loadSampleAppData("sampleuser", "nginx", "sample2@sample.com", "sampleuser-app", true)
		d.loadSampleAppData("maliciousmaintainer", "maliciousapp", "sample3@sample.com", "malicious-app", false) // TODO !! not sure where is the best location for download verification, in the client library or in the cloud? not sure whether I still need this app
	}
}

// TODO !! should be its own object? DatabaseSampleDataSeeder or so?
func (d *DatabaseSampleDataSeeder) loadSampleAppData(username, appname, email, sampleDir string, shouldBeValid bool) {
	err := d.UserRepo.CreateAndValidateUser(&store.RegistrationForm{
		User:     username,
		Password: "password",
		Email:    email,
	})
	if err != nil {
		u.Logger.Error("Failed to create user", tools.UserField, username, deepstack.ErrorField, err)
		os.Exit(1)
	}
	if err = d.AppRepo.CreateApp(username, appname); err != nil {
		u.Logger.Error("Failed to create app", tools.AppField, appname, deepstack.ErrorField, err)
		os.Exit(1)
	}
	appId, err := d.AppRepo.GetAppId(username, appname)
	if err != nil {
		u.Logger.Error("Failed to get app ID", deepstack.ErrorField, err)
		os.Exit(1)
	}
	if err = d.VersionRepo.CreateVersion(appId, "0.0.1",
		tools.GetVersionBytesOfSampleUserApp(sampleDir, username, appname, shouldBeValid)); err != nil {
		u.Logger.Error("Failed to create sample version", deepstack.ErrorField, err)
		os.Exit(1)
	}
}
