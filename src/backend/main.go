package main

import (
	"context"
	"fmt"
	"net/http"
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"
	"ocelot/store/versions"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
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

	deps := WireDependencies()
	fmt.Printf("todo !! temp: %v", deps)

	// TODO !! make this dependent on the config object -> when constructing the config object
	if os.Getenv("USE_MOCK_EMAIL_CLIENT") == "true" {
		u.Logger.Warn("using mock email client, should only be used for testing")
		tools.UseMailMockClient = true
	}

	// TODO !! base config initializer
	err := users.InitializeEnvs()
	if err != nil {
		u.Logger.Error("exiting due to error through env file", deepstack.ErrorField, err)
	}

	// TODO !! database initializer
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
	applyOriginCheckingHandler(mux)
	srv := &http.Server{
		Addr:         ":" + tools.Port,
		Handler:      applyOriginCheckingHandler(mux),
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

// TODO !! not necessary I guess, strict cookie policy should suffice?
func applyOriginCheckingHandler(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := removeSchemeAndPortIfPresent(r.Header.Get("Origin"))
		host := removeSchemeAndPortIfPresent(r.Host)

		if origin == "" || origin == host {
			mux.ServeHTTP(w, r)
		} else {
			u.Logger.Info("request failed since origin header differed from host header", tools.OriginField, origin, tools.HostField, host)
			http.Error(w, "When 'Origin' header is set, it must match host header", http.StatusBadRequest)
			return
		}
	})
}

func removeSchemeAndPortIfPresent(url string) string {
	var newUrl string
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		newUrl = strings.Split(url, "://")[1]
	} else {
		newUrl = url
	}

	if strings.Contains(newUrl, ":") {
		newUrl = strings.Split(newUrl, ":")[0]
	}

	return newUrl
}

type Route struct {
	path    string
	handler http.HandlerFunc
}

type HandlerInitializer struct {
	AppsHandler              *apps.AppsHandler
	VersionsHandler          *versions.VersionsHandler
	DatabaseSampleDataSeeder *DatabaseSampleDataSeeder
	UserHandler              *users.UserHandler
	UserRepo                 users.UserRepository // TODO !! to be removed
}

// TODO !! should be an object which is initialized with the handlers
func (h *HandlerInitializer) InitializeHandlers(mux *http.ServeMux) {
	unprotectedRoutes := []Route{
		{store.LoginPath, h.UserHandler.LoginHandler},
		{store.RegistrationPath, h.UserHandler.RegistrationHandler},
		{store.EmailValidationPath, h.UserHandler.ValidationCodeHandler},

		{store.DownloadPath, h.VersionsHandler.VersionDownloadHandler},
		{store.GetVersionsPath, h.VersionsHandler.GetVersionsHandler},
		{store.SearchAppsPath, h.AppsHandler.SearchForAppsHandler},

		// TODO !! abstract
		{"/api/healthcheck", users.HealthCheckHandler},
	}

	protectedRoutes := []Route{
		{store.AuthCheckPath, h.UserHandler.AuthCheckHandler},
		{store.ChangePasswordPath, h.UserHandler.ChangePasswordHandler},
		{store.DeleteUserPath, h.UserHandler.UserDeleteHandler},
		{store.LogoutPath, h.UserHandler.LogoutHandler},

		{store.VersionUploadPath, h.VersionsHandler.VersionUploadHandler},
		{store.VersionDeletePath, h.VersionsHandler.VersionDeleteHandler},
		{store.AppCreationPath, h.AppsHandler.AppCreationHandler},
		{store.AppGetListPath, h.AppsHandler.AppGetListHandler},
		{store.AppDeletePath, h.AppsHandler.AppDeleteHandler},
	}

	// TODO !! should be called in main
	if tools.Profile == tools.TEST {
		h.UserRepo.WipeDatabase()
		u.Logger.Warn("opening unprotected full data wipe endpoint meant for testing only")
		unprotectedRoutes = append(unprotectedRoutes, Route{store.WipeDataPath, h.UserHandler.WipeData})
		// This user is created to manually test the GUI so that account registration can be skipped to save time.
		sampleUser := "sample"
		// The user may already exist from previous runs. In this case, ignore the error.
		err := h.UserRepo.CreateAndValidateUser(&store.RegistrationForm{
			User:     sampleUser,
			Password: "password",
			Email:    "sample@sample.com",
		})
		if err != nil {
			u.Logger.Debug("Failed to create user - maybe because he already exists, error", tools.UserField, sampleUser, deepstack.ErrorField, err)
		}
		u.Logger.Warn("created user with weak password for manual testing", tools.UserField, sampleUser)
		h.DatabaseSampleDataSeeder.loadSampleAppData("sampleuser", "nginx", "sample2@sample.com", "sampleuser-app", true)
		h.DatabaseSampleDataSeeder.loadSampleAppData("maliciousmaintainer", "maliciousapp", "sample3@sample.com", "malicious-app", false) // TODO !! I think malicious app is no longer needed
	}

	h.registerUnprotectedRoutes(mux, unprotectedRoutes)
	h.registerProtectedRoutes(mux, protectedRoutes)
}

type DatabaseSampleDataSeeder struct {
	AppRepo     apps.AppRepository
	VersionRepo versions.VersionRepository
	UserRepo    users.UserRepository
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

func (h *HandlerInitializer) registerUnprotectedRoutes(mux *http.ServeMux, routes []Route) {
	for _, r := range routes {
		mux.HandleFunc(r.path, r.handler)
	}
}

func (h *HandlerInitializer) registerProtectedRoutes(mux *http.ServeMux, routes []Route) {
	for _, r := range routes {
		mux.Handle(r.path, h.authMiddleware(r.handler))
	}
}

func (h *HandlerInitializer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := h.UserHandler.CheckAuthentication(w, r)
		if err != nil {
			return
		}
		ctx := context.WithValue(r.Context(), tools.UserCtxKey, user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
