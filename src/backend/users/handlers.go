package users

import (
	"fmt"
	"github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
	"net/http"
	"ocelot/store/tools"
	"time"
)

const (
	TestUserWithExpiredCookie          = "expcookietestuser"
	TestUserWithOldButNotExpiredCookie = "oldcookietestuser"
)

var Logger = tools.Logger

func WipeDataHandler(w http.ResponseWriter, r *http.Request) {
	UserRepo.WipeDatabase()
	Logger.Warn("database wipe completed")
	w.WriteHeader(http.StatusOK)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	creds, err := validation.ReadBody[tools.LoginCredentials](w, r)
	if err != nil {
		return
	}

	if !UserRepo.DoesUserExist(creds.User) {
		Logger.Info("user '%s' does not exist", creds.User)
		http.Error(w, "user does not exist", http.StatusNotFound)
		return
	}

	if !UserRepo.IsPasswordCorrect(creds.User, creds.Password) {
		Logger.Info("Password of user '%s' was not correct", creds.User)
		http.Error(w, "incorrect username or password", http.StatusUnauthorized)
		return
	}

	cookie, err := utils.GenerateCookie()
	if err != nil {
		Logger.Error("cookie generation failed: %v", err)
		http.Error(w, "cookie generation failed", http.StatusInternalServerError)
		return
	}

	if tools.Profile == tools.TEST {
		if creds.User == TestUserWithExpiredCookie {
			cookie.Expires = time.Now().UTC().Add(-1 * time.Second)
		} else if creds.User == TestUserWithOldButNotExpiredCookie {
			cookie.Expires = time.Now().UTC().Add(24 * time.Hour)
		}
	}

	err = UserRepo.HashAndSaveCookie(creds.User, cookie.Value, cookie.Expires)
	if err != nil {
		Logger.Error("setting cookie failed: %v", err)
		http.Error(w, "setting cookie failed", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, cookie)
	Logger.Info("user '%s' logged in successfully", creds.User)
	w.WriteHeader(http.StatusOK)
}

func AuthCheckHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	utils.SendJsonResponse(w, tools.UserNameString{Value: user})
}

func UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)

	if !UserRepo.DoesUserExist(user) {
		Logger.Error("user '%s' wanted to delete his account but seems not to exist although authenticated", user)
		http.Error(w, "user does not exist", http.StatusInternalServerError)
		return
	}

	err := UserRepo.DeleteUser(user)
	if err != nil {
		Logger.Error("user '%s' deletion failed", err)
		http.Error(w, "user deletion failed", http.StatusInternalServerError)
		return
	}

	Logger.Info("deleted user: %s", user)
	w.WriteHeader(http.StatusOK)
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)

	form, err := validation.ReadBody[tools.ChangePasswordForm](w, r)
	if err != nil {
		return
	}

	if !UserRepo.DoesUserExist(user) {
		Logger.Warn("somebody tried to change password but user '%s' does not exist", user)
		http.Error(w, "user does not exist", http.StatusNotFound)
		return
	}

	if !UserRepo.IsPasswordCorrect(user, form.OldPassword) {
		Logger.Info("incorrect credentials for user '%s' when trying to change password", user)
		http.Error(w, "incorrect username or password", http.StatusUnauthorized)
		return
	}

	err = UserRepo.ChangePassword(user, form.NewPassword)
	if err != nil {
		Logger.Error("changing password for user '%s' failed: %v", user, err)
		http.Error(w, "error when trying to change password", http.StatusInternalServerError)
		return
	}

	Logger.Info("user '%s' changed his password", user)
	w.WriteHeader(http.StatusOK)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)

	err := UserRepo.Logout(user)
	if err != nil {
		Logger.Error("logout of user '%s' failed: %v", user, err)
		http.Error(w, "logout failed", http.StatusInternalServerError)
		return
	}

	Logger.Info("user '%s' logged out", user)
	w.WriteHeader(http.StatusOK)
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	form, err := validation.ReadBody[tools.RegistrationForm](w, r)
	if err != nil {
		return
	}

	if UserRepo.DoesUserExist(form.User) {
		Logger.Info("user '%s' tried to register but he already exists", form.User)
		http.Error(w, "user already exists", http.StatusConflict)
		return
	}

	if UserRepo.DoesEmailExist(form.Email) {
		Logger.Info("user '%s' tried to register but email '%s' already exists", form.User, form.Email)
		http.Error(w, "email already exists", http.StatusConflict)
		return
	}

	code, err := UserRepo.CreateUser(form)
	if err != nil {
		Logger.Error("user '%s' registration failed: %v", form.User, err)
		http.Error(w, "user registration failed", http.StatusInternalServerError)
		return
	}

	err = sendVerificationEmail(form.Email, code)
	if err != nil {
		Logger.Error("sending verification email failed: %v", err)
		http.Error(w, "sending verification email failed", http.StatusInternalServerError)
		return
	}

	Logger.Info("user wants to register, validation still necessary: " + form.User)
	w.WriteHeader(http.StatusOK)
}

func ValidationCodeHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	code := queryParams.Get("code")

	err := validation.ValidateSecret(code)
	if err != nil {
		tools.HandleInvalidInput(w, err)
		return
	}

	err = UserRepo.ValidateUser(code)
	if err != nil {
		Logger.Error("validation process of user failed: %v", err)
		http.Error(w, "validation process failed", http.StatusBadRequest)
		return
	}

	Logger.Info("user validation code accepted")
	w.WriteHeader(http.StatusOK)
}

func CheckAuthentication(w http.ResponseWriter, r *http.Request) (string, error) {
	Logger.Debug("path: %s", r.URL.Path)
	cookie, err := r.Cookie(tools.CookieName)
	if err != nil {
		Logger.Info("cookie not set in request: %s", err.Error())
		http.Error(w, "cookie not set in request", http.StatusUnauthorized)
		return "", fmt.Errorf("")
	}

	if err = validation.ValidateSecret(cookie.Value); err != nil {
		http.Error(w, "invalid cookie", http.StatusBadRequest)
		return "", fmt.Errorf("")
	}

	user, err := UserRepo.GetUserViaCookie(cookie.Value)
	if err != nil {
		Logger.Info("error when getting cookie of user: %s", err.Error())
		http.Error(w, "cookie not found", http.StatusUnauthorized)
		return "", fmt.Errorf("")
	}

	if UserRepo.IsCookieExpired(cookie.Value) {
		Logger.Warn("user '%s' used an expired cookie'", user)
		http.Error(w, "cookie expired", http.StatusBadRequest)
		return "", fmt.Errorf("")
	}

	newExpirationTime := utils.GetTimeIn30Days()
	err = UserRepo.HashAndSaveCookie(user, cookie.Value, newExpirationTime)
	if err != nil {
		Logger.Error("setting new cookie failed: %v", err)
		http.Error(w, "setting new cookie failed", http.StatusInternalServerError)
		return "", fmt.Errorf("")
	}
	cookie.Expires = newExpirationTime
	// Note: If no path is given, browsers set the default path one level higher than the
	// request path. For example, calling "/a" sets the cookie path to two "/", and calling
	// "/a/b" sets the cookie path to "/a". When updating a cookie, two cookies, the old one
	// and the updated one, with different paths are stored in the browser, causing some
	// requests to fail with "cookie not found".
	cookie.Path = "/"
	cookie.SameSite = http.SameSiteStrictMode
	http.SetCookie(w, cookie)

	return user, nil
}
