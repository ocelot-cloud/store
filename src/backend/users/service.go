package users

import (
	"fmt"
	"net/http"
	"ocelot/store/tools"
	"time"

	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
	"golang.org/x/crypto/bcrypt"
)

var (
	UserDoesNotExistError            = "user does not exist"
	IncorrectUsernameOrPasswordError = "incorrect username or password"
	UserAlreadyExistsError           = "user already exists"
	EmailAlreadyExistsError          = "email already exists"
	InvalidCookieError               = "invalid cookie"
	CookieExpiredError               = "cookie expired"
	CookieNotFoundError              = "cookie not found"
	InvalidInputError                = "invalid input"
)

type UserServiceImpl struct {
	UserRepo                 UserRepository
	Config                   *tools.Config
	EmailVerifier            *tools.EmailVerifierImpl
	EmailClient              *EmailClientImpl
	RegistrationCodeProvider *RegistrationCodeProvider
}

// TODO !! can be made lower case I guess?
func (r *UserServiceImpl) IsPasswordCorrect(userName string, password string) (bool, error) {
	user, err := r.UserRepo.GetUserByName(userName)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	return err == nil, nil
}

// TODO !! cookie expiration time should be real postgres timestamp type
func (r *UserServiceImpl) saveCookie(userName, cookie string, expirationDate time.Time) error {
	user, err := r.UserRepo.GetUserByName(userName)
	if err != nil {
		return err
	}

	hashedCookieValue := u.GetSHA256Hash(cookie)
	user.HashedCookieValue = &hashedCookieValue
	expirationDateString := expirationDate.Format(time.RFC3339)
	user.ExpirationDate = &expirationDateString
	return r.UserRepo.UpdateUser(user)
}

func (r *UserServiceImpl) isCookieExpired(cookie string) (bool, error) {
	hashedCookieValue := u.GetSHA256Hash(cookie)
	user, err := r.UserRepo.GetUserViaCookie(hashedCookieValue)
	if err != nil {
		return true, err
	}

	expirationDate, err := time.Parse(time.RFC3339, *user.ExpirationDate)
	if err != nil {
		return true, u.Logger.NewError(err.Error())
	}
	return time.Now().UTC().After(expirationDate), nil
}

// TODO !! this seems to be a check I should do when uploading and app, directy in the service, so this becomes a non-public method
func (r *UserServiceImpl) IsThereEnoughSpaceToAddVersion(userId, bytesToAdd int) error {
	user, err := r.UserRepo.GetUserById(userId)
	if err != nil {
		return err
	}
	if user.UsedSpaceInBytes+bytesToAdd > tools.MaxStorageSizeInBytes {
		usedStorageInPercent := user.UsedSpaceInBytes * 100 / tools.MaxStorageSizeInBytes
		// TODO !! the "10" should come from a global constant
		msg := fmt.Sprintf(NotEnoughSpacePrefix+", you can't store more then 10MiB of version content, currently used storage in bytes: %d/%d (%d percent)", user.UsedSpaceInBytes, tools.MaxStorageSizeInBytes, usedStorageInPercent)
		// TODO !! this is some information I want to see as user. Maybe keep this a generic message like "data space limit exceeded, see 'status'", introduce a "status" command, showing your account details, including space used absolutely and in percent
		return u.Logger.NewError(msg)
	}
	return nil
}

func (r *UserServiceImpl) WipeDatabase() {
	r.UserRepo.WipeUsers()
	r.EmailVerifier.Clear()
}

func (r *UserServiceImpl) Login(creds *store.LoginCredentials) (*http.Cookie, error) {
	isCorrect, err := r.IsPasswordCorrect(creds.User, creds.Password)
	if err != nil {
		return nil, err
	}

	if !isCorrect {
		return nil, u.Logger.NewError(IncorrectUsernameOrPasswordError)
	}

	// TODO !! test that cookies have expiration date of 7 days; test that cookie is renewed on every authenticated request
	// TODO !! add a unit test in authentication handler that a request with cookie older than 7 days fails
	cookie, err := u.GenerateCookie()
	if err != nil {
		return nil, err
	}

	err = r.saveCookie(creds.User, cookie.Value, cookie.Expires)
	if err != nil {
		return nil, err
	}
	return cookie, nil
}

func (r *UserServiceImpl) RegisterUser(form *store.RegistrationForm) error {
	doesUserExist, err := r.UserRepo.DoesUserExist(form.User)
	if err != nil {
		return err
	}

	if doesUserExist {
		return u.Logger.NewError(UserAlreadyExistsError)
	}

	doesEmailExist, err := r.UserRepo.DoesEmailExist(form.Email)
	if err != nil {
		return err
	}
	if doesEmailExist {
		return u.Logger.NewError(EmailAlreadyExistsError)
	}

	registrationCode, err := r.RegistrationCodeProvider.GenerateCode()
	if err != nil {
		return err
	}
	r.EmailVerifier.Store(registrationCode, form)

	err = r.EmailClient.SendVerificationEmail(form.Email, registrationCode)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserServiceImpl) ValidateUser(code string) error {
	err := validation.ValidateSecret(code)
	if err != nil {
		return err
	}

	form, err := r.EmailVerifier.Load(code)
	if err != nil {
		return err
	}

	err = r.UserRepo.CreateUser(form)
	if err != nil {
		return err
	}
	r.EmailVerifier.Delete(code)
	return nil
}

func (h *UserServiceImpl) CheckAuthentication(cookie *http.Cookie) (*tools.User, *http.Cookie, error) {
	if err := validation.ValidateSecret(cookie.Value); err != nil {
		return nil, nil, u.Logger.NewError(InvalidCookieError)
	}

	hashedCookieValue := u.GetSHA256Hash(cookie.Value)
	user, err := h.UserRepo.GetUserViaCookie(hashedCookieValue)
	if err != nil {
		return nil, nil, err
	}

	isExpired, err := h.isCookieExpired(cookie.Value)
	if err != nil {
		return nil, nil, err
	}
	if isExpired {
		return nil, nil, u.Logger.NewError(CookieExpiredError)
	}

	newExpirationTime := u.GetTimeInSevenDays()
	err = h.saveCookie(user.Name, cookie.Value, newExpirationTime)
	if err != nil {
		return nil, nil, err
	}
	cookie.Expires = newExpirationTime
	// Note: If no path is given, browsers set the default path one level higher than the
	// request path. For example, calling "/a" sets the cookie path to "/", and calling
	// "/a/b" sets the cookie path to "/a". When updating a cookie, two cookies, the old one
	// and the updated one, with different paths are stored in the browser, causing some
	// requests to fail with "cookie not found".
	cookie.Path = "/"
	cookie.SameSite = http.SameSiteStrictMode

	return user, cookie, nil
}

func (r *UserServiceImpl) ChangePassword(user tools.User, form *store.ChangePasswordForm) error {
	isCorrect, err := r.IsPasswordCorrect(user.Name, form.OldPassword)
	if err != nil {
		return err
	}
	if !isCorrect {
		return u.Logger.NewError(IncorrectUsernameOrPasswordError)
	}
	err = r.UserRepo.ChangePassword(user.Id, form.NewPassword)
	if err != nil {
		return err
	}
	return nil
}

// TODO !! does deleting an app free up all space of the versions? -> maybe make it delete versions step by step so used space is calculated correctly?
// TODO feature idea: install an app via direct upload via ocelotcloud web interface -> e.g. for local testing?
