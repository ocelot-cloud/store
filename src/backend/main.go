package main

import (
	"ocelot/store/setup"
	"os"
	"os/exec"

	"github.com/ocelot-cloud/deepstack"
	u "github.com/ocelot-cloud/shared/utils"
)

func main() {
	deps := setup.WireDependencies()
	err := initializeApplication(deps)
	if err != nil {
		u.Logger.Error("exiting due to error during initialization", deepstack.ErrorField, err)
		os.Exit(1)
	}
}

func initializeApplication(deps *setup.InitializerDependencies) error {
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err != nil {
		return u.Logger.NewError(err.Error())
	}
	if err := deps.PathProvider.Initialize(); err != nil {
		return err
	}
	if err := deps.DatabaseProvider.InitializeDatabase(); err != nil {
		return err
	}
	deps.DatabaseSampleDataSeeder.SeedSampleDataForTestMode()
	deps.HandlerInitializer.InitializeHandlers()
	return deps.Server.Run()
}
