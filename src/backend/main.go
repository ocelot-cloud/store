package main

import (
	"context"
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
)

var Logger = tools.Logger

func main() {
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err != nil {
		Logger.Error("docker compose is not installed or not accessible in PATH. Tool is required for docker-compose.yml validation.")
		os.Exit(1)
	}
	if os.Getenv("USE_MOCK_EMAIL_CLIENT") == "true" {
		Logger.Warn("using mock email client, should only be used for testing")
		tools.UseMailMockClient = true
	}
	err := users.InitializeEnvs()
	if err != nil {
		Logger.Error("exiting due to error through env file", deepstack.ErrorField, err)
	}
	tools.InitializeDatabase()
	mux := http.NewServeMux()
	initializeHandlers(mux)
	initializeFrontendResourceDelivery(mux)

	Logger.Info("server starting on port", tools.PortField, tools.Port)
	var handler http.Handler
	if tools.Profile == tools.TEST {
		Logger.Warn("CORS is disabled in test mode")
		handler = GetCorsDisablingHandler(mux)
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
		Logger.Error("Server stopped", deepstack.ErrorField, err)
		os.Exit(1)
	}
}

// TODO !! make the integration tests run against a locally build docker container, so that CORS disabling is not longer needed
func GetCorsDisablingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func applyOriginCheckingHandler(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := removeSchemeAndPortIfPresent(r.Header.Get("Origin"))
		host := removeSchemeAndPortIfPresent(r.Host)

		if origin == "" || origin == host {
			mux.ServeHTTP(w, r)
		} else {
			Logger.Info("request failed since origin header differed from host header", tools.OriginField, origin, tools.HostField, host)
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
		{store.LoginPath, users.LoginHandler},
		{store.DownloadPath, versions.VersionDownloadHandler},
		{store.GetVersionsPath, versions.GetVersionsHandler},
		{store.SearchAppsPath, apps.SearchForAppsHandler},
		{store.RegistrationPath, users.RegistrationHandler},
		{store.EmailValidationPath, users.ValidationCodeHandler},
	}

	protectedRoutes := []Route{
		{store.AuthCheckPath, users.AuthCheckHandler},
		{store.VersionUploadPath, versions.VersionUploadHandler},
		{store.VersionDeletePath, versions.VersionDeleteHandler},
		{store.ChangePasswordPath, users.ChangePasswordHandler},
		{store.AppCreationPath, apps.AppCreationHandler},
		{store.AppGetListPath, apps.AppGetListHandler},
		{store.AppDeletePath, apps.AppDeleteHandler},
		{store.DeleteUserPath, users.UserDeleteHandler},
		{store.LogoutPath, users.LogoutHandler},
	}

	if tools.Profile == tools.TEST {
		users.UserRepo.WipeDatabase()
		Logger.Warn("opening unprotected full data wipe endpoint meant for testing only")
		unprotectedRoutes = append(unprotectedRoutes, Route{store.WipeDataPath, users.WipeDataHandler})
		// This user is created to manually test the GUI so that account registration can be skipped to save time.
		sampleUser := "sample"
		// The user may already exist from previous runs. In this case, ignore the error.
		err := users.CreateAndValidateUser(&store.RegistrationForm{
			User:     sampleUser,
			Password: "password",
			Email:    "sample@sample.com",
		})
		if err != nil {
			Logger.Debug("Failed to create user - maybe because he already exists, error", tools.UserField, sampleUser, deepstack.ErrorField, err)
		}
		Logger.Warn("created user with weak password for manual testing", tools.UserField, sampleUser)
		loadSampleAppData("sampleuser", "nginx", "sample2@sample.com", "sampleuser-app", true)
		loadSampleAppData("maliciousmaintainer", "maliciousapp", "sample3@sample.com", "malicious-app", false)
	}

	registerUnprotectedRoutes(mux, unprotectedRoutes)
	registerProtectedRoutes(mux, protectedRoutes)
}

func loadSampleAppData(username, appname, email, sampleDir string, shouldBeValid bool) {
	err := users.CreateAndValidateUser(&store.RegistrationForm{
		User:     username,
		Password: "password",
		Email:    email,
	})
	if err != nil {
		Logger.Error("Failed to create user", tools.UserField, username, deepstack.ErrorField, err)
		os.Exit(1)
	}
	if err = apps.AppRepo.CreateApp(username, appname); err != nil {
		Logger.Error("Failed to create app", tools.AppField, appname, deepstack.ErrorField, err)
		os.Exit(1)
	}
	appId, err := apps.AppRepo.GetAppId(username, appname)
	if err != nil {
		Logger.Error("Failed to get app ID", deepstack.ErrorField, err)
		os.Exit(1)
	}
	if err = versions.VersionRepo.CreateVersion(appId, "0.0.1",
		tools.GetVersionBytesOfSampleUserApp(sampleDir, username, appname, shouldBeValid)); err != nil {
		Logger.Error("Failed to create sample version", deepstack.ErrorField, err)
		os.Exit(1)
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
			Logger.Debug("Serving index.html for SPA route/path", tools.UrlPathField, r.URL.Path)
			http.ServeFile(w, r, "./dist/index.html")
			return
		}

		// If the request is for a static file or if the file exists, serve it directly.
		// This handles requests for JS, CSS, images, etc.
		Logger.Debug("Serving static content", tools.UrlPathField, r.URL.Path)
		http.FileServer(http.Dir("./dist")).ServeHTTP(w, r)
	}))
}
