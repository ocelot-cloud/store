package setup

import (
	"context"
	"net/http"
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"
	"ocelot/store/versions"

	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
)

type Route struct {
	path    string
	handler http.HandlerFunc
}

type HandlerInitializer struct {
	AppsHandler     *apps.AppsHandler
	VersionsHandler *versions.VersionsHandler
	UserHandler     *users.UserHandler
	Config          *tools.Config
	Mux             *http.ServeMux
}

func (h *HandlerInitializer) InitializeHandlers() {
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

	if h.Config.OpenWipeEndpoint {
		u.Logger.Warn("opening unprotected full data wipe endpoint meant for testing only")
		unprotectedRoutes = append(unprotectedRoutes, Route{store.WipeDataPath, h.UserHandler.WipeData})
	}

	h.registerUnprotectedRoutes(unprotectedRoutes)
	h.registerProtectedRoutes(protectedRoutes)
}

func (h *HandlerInitializer) registerUnprotectedRoutes(routes []Route) {
	for _, r := range routes {
		h.Mux.HandleFunc(r.path, r.handler)
	}
}

func (h *HandlerInitializer) registerProtectedRoutes(routes []Route) {
	for _, r := range routes {
		h.Mux.Handle(r.path, h.authMiddleware(r.handler))
	}
}

func (h *HandlerInitializer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := h.UserHandler.CheckAuthentication(w, r)
		if err != nil {
			return
		}
		ctx := context.WithValue(r.Context(), tools.UserCtxKey, *user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
