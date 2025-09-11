package main

import (
	"net/http"
	"ocelot/store/setup"
	"ocelot/store/tools"
	"os"
	"os/exec"
	"time"

	"github.com/ocelot-cloud/deepstack"
	u "github.com/ocelot-cloud/shared/utils"
)

// TODO !! way to long, simply call wire here and initialize the modules, rest is to be extracted
// TODO !! replace errors with deepstack approach
// TODO !! shift logic from handlers and repos to services, simplify repos to CRUD

func main() {
	// TODO !! tool installation checker
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err != nil {
		u.Logger.Error("docker compose is not installed or not accessible in PATH. Tool is required for docker-compose.yml validation.")
		os.Exit(1)
	}

	deps := setup.WireDependencies()
	err := initializeApplication(deps)
	// TODO !! all other lines of the main function should be move into this function, and called from "deps"
	err = deps.DatabaseProvider.InitializeDatabase()
	if err != nil {
		u.Logger.Error("exiting due to error through database", deepstack.ErrorField, err)
		os.Exit(1)
	}

	// TODO !! handler initializer
	mux := http.NewServeMux()
	// TODO !! mux should be injected internally I guess and not via main?
	deps.HandlerInitializer.InitializeHandlers(mux)

	// TODO !! server.run()
	srv := &http.Server{
		Addr:         ":" + tools.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	u.Logger.Info("server starting on port", tools.PortField, tools.Port)
	err = srv.ListenAndServe()
	if err != nil {
		u.Logger.Error("Server stopped", deepstack.ErrorField, err)
		os.Exit(1)
	}
}

func initializeApplication(deps *setup.InitializerDependencies) interface{} {
	err := deps.PathProvider.Initialize()
	if err != nil {
		return err
	}
	return nil
}
