package main

import (
	"context"
	"github.com/ocelot-cloud/shared/utils"
	"net/http"
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"
	"ocelot/store/versions"
	"os"
	"os/exec"
	"strings"
	"time"
)

func init() {
	tools.Logger = utils.ProvideLogger(os.Getenv("LOG_LEVEL"))
}

func main() {
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err != nil {
		tools.Logger.Fatal("docker compose is not installed or not accessible in PATH")
	}
	if os.Getenv("USE_MOCK_EMAIL_CLIENT") == "true" {
		tools.Logger.Warn("using mock email client, should only be used for testing")
		tools.UseMailMockClient = true
	}
	if tools.Profile == tools.TEST {
		tools.Logger.Info("profile is: TEST")
	} else if tools.Profile == tools.PROD {
		tools.Logger.Info("profile is: PROD")
	} else {
		tools.Logger.Fatal("unknown profile: %d", tools.Profile)
	}
	err := users.InitializeEnvs()
	if err != nil {
		tools.Logger.Fatal("exiting due to error through env file: %v", err)
	}
	tools.InitializeDatabase()
	mux := http.NewServeMux()
	initializeHandlers(mux)
	initializeFrontendResourceDelivery(mux)

	tools.Logger.Info("server starting on port %s", tools.Port)
	var handler http.Handler
	if tools.Profile == tools.TEST {
		tools.Logger.Warn("CORS is disabled in test mode")
		handler = utils.GetCorsDisablingHandler(mux)
	} else {
		handler = applyOriginCheckingHandler(mux)
	}
	srv := &http.Server{
		Addr:         ":" + tools.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	err = srv.ListenAndServe()
	if err != nil {
		tools.Logger.Error("Server stopped: %v", err)
		os.Exit(1)
	}
}

func applyOriginCheckingHandler(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := removeSchemeAndPortIfPresent(r.Header.Get("Origin"))
		host := removeSchemeAndPortIfPresent(r.Host)

		if origin == "" || origin == host {
			mux.ServeHTTP(w, r)
		} else {
			tools.Logger.Info("request failed since origin header '%s' differed from host header '%s'", origin, host)
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

func initializeHandlers(mux *http.ServeMux) {
	unprotectedRoutes := []Route{
		{tools.LoginPath, users.LoginHandler},
		{tools.DownloadPath, versions.VersionDownloadHandler},
		{tools.GetVersionsPath, versions.GetVersionsHandler},
		{tools.SearchAppsPath, apps.SearchForAppsHandler},
		{tools.RegistrationPath, users.RegistrationHandler},
		{tools.EmailValidationPath, users.ValidationCodeHandler},
	}

	protectedRoutes := []Route{
		{tools.AuthCheckPath, users.AuthCheckHandler},
		{tools.VersionUploadPath, versions.VersionUploadHandler},
		{tools.VersionDeletePath, versions.VersionDeleteHandler},
		{tools.ChangePasswordPath, users.ChangePasswordHandler},
		{tools.AppCreationPath, apps.AppCreationHandler},
		{tools.AppGetListPath, apps.AppGetListHandler},
		{tools.AppDeletePath, apps.AppDeleteHandler},
		{tools.DeleteUserPath, users.UserDeleteHandler},
		{tools.LogoutPath, users.LogoutHandler},
	}

	if tools.Profile == tools.TEST {
		users.UserRepo.WipeDatabase()
		tools.Logger.Warn("opening unprotected full data wipe endpoint meant for testing only")
		unprotectedRoutes = append(unprotectedRoutes, Route{tools.WipeDataPath, users.WipeDataHandler})
		// This user is created to manually test the GUI so that account registration can be skipped to save time.
		sampleUser := "sample"
		// The user may already exist from previous runs. In this case, ignore the error.
		err := users.CreateAndValidateUser(&tools.RegistrationForm{
			User:     sampleUser,
			Password: "password",
			Email:    "sample@sample.com",
		})
		if err != nil {
			tools.Logger.Debug("Failed to create user '%s' - maybe because he already exists, error: %v.", sampleUser, err)
		}
		tools.Logger.Warn("created '%s' user with weak password for manual testing", sampleUser)
		loadSampleApp()
	}

	registerUnprotectedRoutes(mux, unprotectedRoutes)
	registerProtectedRoutes(mux, protectedRoutes)
}

// Creates a sample app that can be downloaded from the cloud for testing.
func loadSampleApp() {
	tools.Logger.Warn("loading sample app 'nginxdefault' into database for testing")
	sampleUser := "sampleuser"
	sampleApp := "nginxdefault"
	err := users.CreateAndValidateUser(&tools.RegistrationForm{
		User:     sampleUser,
		Password: "password",
		Email:    "sample2@sample.com",
	})
	if err != nil {
		tools.Logger.Fatal("Failed to create '%s' user: %v.", sampleUser, err)
	}
	err = apps.AppRepo.CreateApp(sampleUser, sampleApp)
	if err != nil {
		tools.Logger.Fatal("Failed to create '%s' app: %v.", sampleApp, err)
	}

	appId, err := apps.AppRepo.GetAppId(sampleUser, sampleApp)
	if err != nil {
		tools.Logger.Fatal("Failed to get app ID: %v", err)
	}
	err = versions.VersionRepo.CreateVersion(appId, "0.0.1", tools.GetValidVersionBytes())
	if err != nil {
		tools.Logger.Fatal("Failed to create sample version: %v", err)
	}
}

func registerUnprotectedRoutes(mux *http.ServeMux, routes []Route) {
	for _, r := range routes {
		mux.HandleFunc(r.path, r.handler)
	}
}

func registerProtectedRoutes(mux *http.ServeMux, routes []Route) {
	for _, r := range routes {
		mux.Handle(r.path, authMiddleware(r.handler))
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := users.CheckAuthentication(w, r)
		if err != nil {
			return
		}
		ctx := context.WithValue(r.Context(), tools.UserCtxKey, user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func initializeFrontendResourceDelivery(mux *http.ServeMux) {
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Attempt to open the requested file within the ./dist directory.
		_, err := http.Dir("./dist").Open(r.URL.Path)

		// If the requested file does not exist (err is not nil) and the path does not seem to refer to
		// a static file (i.e. no dot extension like ".css"), then serve index.html. This caters to SPA routing needs,
		// allowing frontend routes to be handled by index.html.
		// This means that users can directly access pages with paths such as "example.com/some/path".
		if err != nil && !strings.Contains(r.URL.Path, ".") {
			tools.Logger.Debug("Serving index.html for SPA route: %s", r.URL.Path)
			http.ServeFile(w, r, "./dist/index.html")
			return
		}

		// If the request is for a static file or if the file exists, serve it directly.
		// This handles requests for JS, CSS, images, etc.
		tools.Logger.Debug("Serving static content at '%s'", r.URL.Path)
		http.FileServer(http.Dir("./dist")).ServeHTTP(w, r)
	}))
}
