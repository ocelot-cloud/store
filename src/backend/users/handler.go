package users

import (
	"net/http"
	"ocelot/store/tools"

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
		u.WriteResponseError(w, u.MapOf(UserDoesNotExistError, IncorrectUsernameOrPasswordError), err)
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
		u.WriteResponseError(w, nil, err)
		return
	}
}

func (h *UserHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	form, err := validation.ReadBody[store.ChangePasswordForm](w, r)
	if err != nil {
		return
	}
	err = h.UserService.ChangePassword(user, form)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(IncorrectUsernameOrPasswordError), err)
		return
	}
}

func (h *UserHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	user := tools.GetUserFromContext(r)
	err := h.UserRepo.Logout(user.Id)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
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

func (h *UserHandler) ValidationCodeHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	code := queryParams.Get("code")
	err := h.UserService.ValidateUser(code)
	if err != nil {
		u.WriteResponseError(w, u.MapOf(InvalidInputError), err)
		return
	}
}
