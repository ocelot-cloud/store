package users

import (
	"fmt"
	"net/http"
	"ocelot/store/tools"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
)

type UserHandler struct {
	UserRepo    UserRepository
	EmailClient *EmailClientImpl
	Config      *tools.Config
	UserService *UserServiceImpl
}

func (h *UserHandler) WipeData(w http.ResponseWriter, r *http.Request) {
	h.UserService.WipeDatabase()
	u.Logger.Warn("database wipe completed")
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	creds, err := validation.ReadBody[store.LoginCredentials](w, r)
	if err != nil {
		return
	}
	cookie, err := h.UserService.Login(creds)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(UserDoesNotExistError, IncorrectUsernameAndPasswordError), err)
		return
	}
	http.SetCookie(w, cookie)
}

// TODO !! Should we send more info of user object? e.g. cookie expiration time (for testing)?
func (h *UserHandler) AuthCheckHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	u.SendJsonResponse(w, store.UserNameString{Value: user.Name})
}

func (h *UserHandler) UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	err := h.UserRepo.DeleteUser(user.Name)
	if err != nil {
		u.Logger.Error("user deletion failed", tools.UserField, err)
		http.Error(w, "user deletion failed", http.StatusBadRequest)
		return
	}

	u.Logger.Info("deleted user", tools.UserField, user)
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	form, err := validation.ReadBody[store.ChangePasswordForm](w, r)
	if err != nil {
		return
	}

	isCorrect, err := h.UserService.IsPasswordCorrect(user.Name, form.OldPassword)
	if err != nil {
		u.Logger.Error("checking password of user failed", deepstack.ErrorField, err)
		http.Error(w, "error when checking password", http.StatusBadRequest)
		return
	}
	if !isCorrect {
		u.Logger.Info("incorrect credentials for user when trying to change password", tools.UserField, user)
		http.Error(w, IncorrectUsernameAndPasswordError, http.StatusBadRequest)
		return
	}

	err = h.UserRepo.ChangePassword(user.Id, form.NewPassword)
	if err != nil {
		u.Logger.Error("changing password for user failed", tools.UserField, user, deepstack.ErrorField, err)
		http.Error(w, "error when trying to change password", http.StatusBadRequest)
		return
	}

	u.Logger.Info("user changed his password", tools.UserField, user)
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	err := h.UserRepo.Logout(user.Name)
	if err != nil {
		u.Logger.Error("logout of user failed", tools.UserField, user, deepstack.ErrorField, err)
		http.Error(w, "logout failed", http.StatusBadRequest)
		return
	}

	u.Logger.Info("user logged out", tools.UserField, user)
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	form, err := validation.ReadBody[store.RegistrationForm](w, r)
	if err != nil {
		return
	}
	err = h.UserService.RegisterUser(form)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(UserAlreadyExistsError, EmailAlreadyExistsError), err)
		return
	}
}

type healthInfo struct {
	Status string `json:"status"`
}

// TODO !! find better location
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	u.SendJsonResponse(w, healthInfo{Status: "ok"})
}

func (h *UserHandler) ValidationCodeHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	code := queryParams.Get("code")

	err := validation.ValidateSecret(code)
	if err != nil {
		u.WriteResponseError(w, u.MapOf("invalid input"), err)
		return
	}

	err = h.UserService.ValidateUserViaRegistrationCode(code)
	if err != nil {
		u.Logger.Error("validation process of user failed", deepstack.ErrorField, err)
		http.Error(w, "validation process failed", http.StatusBadRequest)
		return
	}

	u.Logger.Info("user validation code accepted")
	w.WriteHeader(http.StatusOK)
}

// TODO !! fmt.Errorf("") looks as if it should be refctored away?
func (h *UserHandler) CheckAuthentication(w http.ResponseWriter, r *http.Request) (*tools.User, error) {
	u.Logger.Debug("checking authentication", tools.UrlPathField, r.URL.Path)
	cookie, err := r.Cookie(tools.CookieName)
	if err != nil {
		u.Logger.Info("cookie not set in request", deepstack.ErrorField, err)
		http.Error(w, "cookie not set in request", http.StatusBadRequest)
		return nil, fmt.Errorf("")
	}

	if err = validation.ValidateSecret(cookie.Value); err != nil {
		http.Error(w, "invalid cookie", http.StatusBadRequest)
		return nil, fmt.Errorf("")
	}

	hashedCookieValue := u.GetSHA256Hash(cookie.Value)
	user, err := h.UserRepo.GetUserViaCookie(hashedCookieValue)
	if err != nil {
		u.Logger.Info("error when getting cookie of user", deepstack.ErrorField, err)
		http.Error(w, "cookie not found", http.StatusBadRequest)
		return nil, fmt.Errorf("")
	}

	isExpired, err := h.UserService.IsCookieExpired(cookie.Value)
	if err != nil {
		u.Logger.Error("checking if cookie is expired failed", deepstack.ErrorField, err)
		http.Error(w, "error when checking if cookie is expired", http.StatusBadRequest)
		return nil, fmt.Errorf("")
	}
	if isExpired {
		u.Logger.Warn("user used an expired cookie", tools.UserField, user)
		http.Error(w, "cookie expired", http.StatusBadRequest)
		return nil, fmt.Errorf("")
	}

	newExpirationTime := u.GetTimeInSevenDays()
	err = h.UserService.SaveCookie(user.Name, cookie.Value, newExpirationTime)
	if err != nil {
		u.Logger.Error("setting new cookie failed", deepstack.ErrorField, err)
		http.Error(w, "setting new cookie failed", http.StatusBadRequest)
		return nil, fmt.Errorf("")
	}
	cookie.Expires = newExpirationTime
	// Note: If no path is given, browsers set the default path one level higher than the
	// request path. For example, calling "/a" sets the cookie path to "/", and calling
	// "/a/b" sets the cookie path to "/a". When updating a cookie, two cookies, the old one
	// and the updated one, with different paths are stored in the browser, causing some
	// requests to fail with "cookie not found".
	cookie.Path = "/"
	cookie.SameSite = http.SameSiteStrictMode
	http.SetCookie(w, cookie)

	return user, nil
}
