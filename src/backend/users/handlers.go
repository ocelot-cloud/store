package users

import (
	"fmt"
	"net/http"
	"ocelot/store/tools"
	"time"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
)

const (
	TestUserWithExpiredCookie          = "expcookietestuser"
	TestUserWithOldButNotExpiredCookie = "oldcookietestuser"
)

type UserHandler struct {
	UserRepo    UserRepository
	EmailClient *EmailClient
	Config      *tools.Config
}

func (h *UserHandler) WipeData(w http.ResponseWriter, r *http.Request) {
	h.UserRepo.WipeDatabase()
	u.Logger.Warn("database wipe completed")
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	creds, err := validation.ReadBody[store.LoginCredentials](w, r)
	if err != nil {
		return
	}

	if !h.UserRepo.DoesUserExist(creds.User) {
		u.Logger.Info("user does not exist", tools.UserField, creds.User)
		http.Error(w, "user does not exist", http.StatusBadRequest)
		return
	}

	if !h.UserRepo.IsPasswordCorrect(creds.User, creds.Password) {
		u.Logger.Info("Password of user was not correct", tools.UserField, creds.User)
		http.Error(w, "incorrect username or password", http.StatusBadRequest)
		return
	}

	cookie, err := u.GenerateCookie()
	if err != nil {
		u.Logger.Error("cookie generation failed", deepstack.ErrorField, err)
		http.Error(w, "cookie generation failed", http.StatusBadRequest)
		return
	}

	if h.Config.UseSpecialExpiration {
		// TODO !! I find this approach very ugly, should be refactored somehow -> when logging in, return user including his expiration data of cookie and assert this instead
		if creds.User == TestUserWithExpiredCookie {
			cookie.Expires = time.Now().UTC().Add(-1 * time.Second)
		} else if creds.User == TestUserWithOldButNotExpiredCookie {
			cookie.Expires = time.Now().UTC().Add(24 * time.Hour)
		}
	}

	err = h.UserRepo.HashAndSaveCookie(creds.User, cookie.Value, cookie.Expires)
	if err != nil {
		u.Logger.Error("setting cookie failed", deepstack.ErrorField, err)
		http.Error(w, "setting cookie failed", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, cookie)
	u.Logger.Info("user logged in successfully", tools.UserField, creds.User)
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) AuthCheckHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	u.SendJsonResponse(w, store.UserNameString{Value: user})
}

func (h *UserHandler) UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)

	if !h.UserRepo.DoesUserExist(user) {
		u.Logger.Error("user wanted to delete his account but seems not to exist although authenticated", tools.UserField, user)
		http.Error(w, "user does not exist", http.StatusBadRequest)
		return
	}

	err := h.UserRepo.DeleteUser(user)
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

	if !h.UserRepo.DoesUserExist(user) {
		u.Logger.Warn("somebody tried to change password but user does not exist", tools.UserField, user)
		http.Error(w, "user does not exist", http.StatusBadRequest)
		return
	}

	if !h.UserRepo.IsPasswordCorrect(user, form.OldPassword) {
		u.Logger.Info("incorrect credentials for user when trying to change password", tools.UserField, user)
		http.Error(w, "incorrect username or password", http.StatusBadRequest)
		return
	}

	err = h.UserRepo.ChangePassword(user, form.NewPassword)
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

	err := h.UserRepo.Logout(user)
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

	if h.UserRepo.DoesUserExist(form.User) {
		u.Logger.Info("user tried to register but he already exists", tools.UserField, form.User)
		http.Error(w, "user already exists", http.StatusBadRequest)
		return
	}

	if h.UserRepo.DoesEmailExist(form.Email) {
		u.Logger.Info("user tried to register but email already exists", tools.UserField, form.User, tools.EmailField, form.Email)
		http.Error(w, "email already exists", http.StatusBadRequest)
		return
	}

	code, err := h.UserRepo.CreateUserAndReturnRegistrationCode(form)
	if err != nil {
		u.Logger.Error("user registration failed", tools.UserField, form.User, deepstack.ErrorField, err)
		http.Error(w, "user registration failed", http.StatusBadRequest)
		return
	}

	err = h.EmailClient.SendVerificationEmail(form.Email, code)
	if err != nil {
		u.Logger.Error("sending verification email failed", deepstack.ErrorField, err)
		http.Error(w, "sending verification email failed", http.StatusBadRequest)
		return
	}

	u.Logger.Info("user wants to register, validation still necessary", tools.UserField, form.User)
	w.WriteHeader(http.StatusOK)
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

	err = h.UserRepo.ValidateUserViaRegistrationCode(code)
	if err != nil {
		u.Logger.Error("validation process of user failed", deepstack.ErrorField, err)
		http.Error(w, "validation process failed", http.StatusBadRequest)
		return
	}

	u.Logger.Info("user validation code accepted")
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) CheckAuthentication(w http.ResponseWriter, r *http.Request) (string, error) {
	u.Logger.Debug("checking authentication", tools.UrlPathField, r.URL.Path)
	cookie, err := r.Cookie(tools.CookieName)
	if err != nil {
		u.Logger.Info("cookie not set in request", deepstack.ErrorField, err)
		http.Error(w, "cookie not set in request", http.StatusBadRequest)
		return "", fmt.Errorf("")
	}

	if err = validation.ValidateSecret(cookie.Value); err != nil {
		http.Error(w, "invalid cookie", http.StatusBadRequest)
		return "", fmt.Errorf("")
	}

	user, err := h.UserRepo.GetUserViaCookie(cookie.Value)
	if err != nil {
		u.Logger.Info("error when getting cookie of user", deepstack.ErrorField, err)
		http.Error(w, "cookie not found", http.StatusBadRequest)
		return "", fmt.Errorf("")
	}

	if h.UserRepo.IsCookieExpired(cookie.Value) {
		u.Logger.Warn("user used an expired cookie", tools.UserField, user)
		http.Error(w, "cookie expired", http.StatusBadRequest)
		return "", fmt.Errorf("")
	}

	newExpirationTime := u.GetTimeInSevenDays()
	err = h.UserRepo.HashAndSaveCookie(user, cookie.Value, newExpirationTime)
	if err != nil {
		u.Logger.Error("setting new cookie failed", deepstack.ErrorField, err)
		http.Error(w, "setting new cookie failed", http.StatusBadRequest)
		return "", fmt.Errorf("")
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
