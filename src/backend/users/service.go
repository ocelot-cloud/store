package users

import (
	"crypto/rand"
	"encoding/hex"
	"ocelot/store/tools"
	"time"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceImpl struct {
	UserRepo      UserRepository
	Config        *tools.Config
	EmailVerifier *tools.EmailVerifierImpl
}

func (r *UserServiceImpl) CreateUserAndReturnRegistrationCode(form *store.RegistrationForm) (string, error) {
	var key string
	// TODO !! quite implicit logic. Maybe a better option to say in test mode, when we create a user, his account needs no code for validation
	if r.Config.UseSampleDataForTesting {
		// TODO static sample key for testing, use from shared module
		key = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	} else {
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			u.Logger.Error("Failed to generate cookie", deepstack.ErrorField, err)
			return "", err
		}
		key = hex.EncodeToString(randomBytes)
	}
	u.Logger.Info("adding user to validation list", tools.UserField, form.User)
	r.EmailVerifier.Store(key, form)
	return key, nil
}

func (r *UserServiceImpl) ValidateUserViaRegistrationCode(code string) error {
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

func (r *UserServiceImpl) IsPasswordCorrect(userName string, password string) (bool, error) {
	user, err := r.UserRepo.GetUserByName(userName)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	return err == nil, nil
}

// TODO !! cookie expiration time should be real postgres timestamp type
func (r *UserServiceImpl) SaveCookie(userName, cookie string, expirationDate time.Time) error {
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
