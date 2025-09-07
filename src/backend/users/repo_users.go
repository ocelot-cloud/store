package users

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"ocelot/store/tools"
	"sync"
	"time"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	"github.com/ocelot-cloud/shared/utils"
	"golang.org/x/crypto/bcrypt"
)

var NotEnoughSpacePrefix = "not enough space"

func (u *UserRepositoryImpl) IsThereEnoughSpaceToAddVersion(user string, bytesToAdd int) error {
	bytesUsed, err := UserRepo.GetUsedSpaceInBytes(user)
	if err != nil {
		Logger.Error("checking space failed", deepstack.ErrorField, err)
		return errors.New("checking space failed")
	}
	if bytesUsed+bytesToAdd > tools.MaxStorageSize {
		Logger.Info("user tried to upload version, but storage limit would be exceeded", tools.UserField, user)
		usedStorageInPercent := bytesUsed * 100 / tools.MaxStorageSize
		msg := fmt.Sprintf(NotEnoughSpacePrefix+", you can't store more then 10MiB of version content, currently used storage in bytes: %d/%d (%d percent)", bytesUsed, tools.MaxStorageSize, usedStorageInPercent)
		return errors.New(msg)
	}
	return nil
}

func (u *UserRepositoryImpl) IsPasswordCorrect(user string, password string) bool {
	var hashedPassword string
	err := tools.Db.QueryRow("SELECT hashed_password FROM users WHERE user_name = $1", user).Scan(&hashedPassword)
	if err != nil {
		Logger.Error("Failed to fetch hashed password", deepstack.ErrorField, err)
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (u *UserRepositoryImpl) DoesUserExist(user string) bool {
	var exists bool
	err := tools.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_name = $1)", user).Scan(&exists)
	if err != nil {
		Logger.Error("Failed to check user existence", deepstack.ErrorField, err)
		return false
	}
	return exists
}

func (u *UserRepositoryImpl) CreateUser(form *store.RegistrationForm) (string, error) {
	var key string
	if tools.UseMailMockClient {
		// TODO static sample key for testing, use from shared module
		key = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	} else {
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			Logger.Error("Failed to generate cookie", deepstack.ErrorField, err)
			return "", err
		}
		key = hex.EncodeToString(randomBytes)
	}
	Logger.Info("adding user to validation list", tools.UserField, form.User)
	tools.WaitingForEmailVerificationList.Store(key, form)
	return key, nil
}

func (u *UserRepositoryImpl) ValidateUser(code string) error {
	value, ok := tools.WaitingForEmailVerificationList.Load(code)
	if !ok {
		return fmt.Errorf("code not found")
	}

	form, ok := value.(*store.RegistrationForm)
	if !ok {
		Logger.Error("Invalid type for registration form")
		return fmt.Errorf("invalid type for registration form")
	}

	hashedPassword, err := utils.SaltAndHash(form.Password)
	if err != nil {
		Logger.Error("Failed to hash password", deepstack.ErrorField, err)
		return fmt.Errorf("failed to hash password")
	}
	_, err = tools.Db.Exec("INSERT INTO users (user_name, email, hashed_password, used_space) VALUES ($1, $2, $3, $4)", form.User, form.Email, hashedPassword, 0)
	if err != nil {
		Logger.Error("Failed to create user", deepstack.ErrorField, err)
		return fmt.Errorf("failed to create user")
	}
	tools.WaitingForEmailVerificationList.Delete(code)
	return nil
}

func (u *UserRepositoryImpl) DeleteUser(user string) error {
	if !u.DoesUserExist(user) {
		Logger.Info("User does not exist", deepstack.ErrorField, user)
		return fmt.Errorf("user does not exist")
	}

	_, err := tools.Db.Exec("DELETE FROM users WHERE user_name = $1", user)
	if err != nil {
		Logger.Error("Failed to delete user", deepstack.ErrorField, err)
		return fmt.Errorf("failed to delete user")
	}

	return nil
}

func (u *UserRepositoryImpl) HashAndSaveCookie(user string, cookie string, expirationDate time.Time) error {
	hashedCookieValue := utils.GetSHA256Hash(cookie)

	_, err := tools.Db.Exec("UPDATE users SET hashed_cookie_value = $1, expiration_date = $2 WHERE user_name = $3", hashedCookieValue, expirationDate.Format(time.RFC3339), user)
	if err != nil {
		Logger.Error("Failed to hash and save cookie", deepstack.ErrorField, err)
		return fmt.Errorf("failed to hash and save cookie")
	}
	return nil
}

func (u *UserRepositoryImpl) IsCookieExpired(cookie string) bool {
	hashedCookieValue := utils.GetSHA256Hash(cookie)

	var expirationDateStr string
	err := tools.Db.QueryRow("SELECT expiration_date FROM users WHERE hashed_cookie_value = $1", hashedCookieValue).Scan(&expirationDateStr)
	if err != nil {
		Logger.Error("Failed to fetch expiration date", deepstack.ErrorField, err)
		return true
	} else if expirationDateStr == "" {
		return true
	}

	expirationDate, err := time.Parse(time.RFC3339, expirationDateStr)
	if err != nil {
		Logger.Error("Failed to parse expiration date", deepstack.ErrorField, err)
		return true
	}

	return time.Now().UTC().After(expirationDate)
}

func (u *UserRepositoryImpl) GetUserViaCookie(cookie string) (string, error) {
	if cookie == "" {
		Logger.Error("Cookie not set in request")
		return "", fmt.Errorf("cookie not set in request")
	}

	hashedCookieValue := utils.GetSHA256Hash(cookie)

	var user string
	err := tools.Db.QueryRow("SELECT user_name FROM users WHERE hashed_cookie_value = $1", hashedCookieValue).Scan(&user)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			Logger.Info("Cookie not found")
			return "", fmt.Errorf("cookie not found")
		} else {
			Logger.Error("Failed to fetch user", deepstack.ErrorField, err)
			return "", fmt.Errorf("failed to fetch user")
		}
	}

	return user, nil
}

func (u *UserRepositoryImpl) ChangePassword(user string, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		Logger.Error("Failed to hash password", deepstack.ErrorField, err)
		return fmt.Errorf("failed to hash password")
	}

	_, err = tools.Db.Exec("UPDATE users SET hashed_password = $1 WHERE user_name = $2", hashedPassword, user)
	if err != nil {
		Logger.Error("Failed to change password", deepstack.ErrorField, err)
		return fmt.Errorf("failed to change password")
	}

	return nil
}

func (u *UserRepositoryImpl) WipeDatabase() {
	_, err := tools.Db.Exec("DELETE FROM users WHERE user_name != 'sample'")
	if err != nil {
		Logger.Error("Failed to wipe database", deepstack.ErrorField, err)
	}
	tools.WaitingForEmailVerificationList = sync.Map{}
}

func (u *UserRepositoryImpl) GetUsedSpaceInBytes(user string) (int, error) {
	var usedSpace int
	err := tools.Db.QueryRow(`SELECT used_space FROM users WHERE user_name = $1`, user).Scan(&usedSpace)
	if err != nil {
		Logger.Error("Failed to get used space", deepstack.ErrorField, err)
		return 0, fmt.Errorf("failed to get used space")
	}
	return usedSpace, nil
}

func (u *UserRepositoryImpl) Logout(user string) error {
	_, err := tools.Db.Exec("UPDATE users SET hashed_cookie_value = $1, expiration_date = $2 WHERE user_name = $3", nil, nil, user)
	if err != nil {
		Logger.Error("failed to logout", deepstack.ErrorField, err)
		return errors.New("failed to logout")
	}
	return nil
}

func CreateAndValidateUser(form *store.RegistrationForm) error {
	code, err := UserRepo.CreateUser(form)
	if err != nil {
		return err
	}
	err = UserRepo.ValidateUser(code)
	return err
}

func (u *UserRepositoryImpl) DoesEmailExist(email string) bool {
	var exists bool
	err := tools.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		Logger.Error("Failed to check email existence", deepstack.ErrorField, err)
		return false
	}
	return exists
}

type UserRepositoryImpl struct{}

// TODO !! global var
var UserRepo UserRepository = &UserRepositoryImpl{}

// TODO !! simplify to CRUD operations, rest should be handle by a service
type UserRepository interface {
	CreateUser(form *store.RegistrationForm) (string, error)
	ValidateUser(code string) error
	DoesUserExist(user string) bool
	DoesEmailExist(email string) bool
	DeleteUser(user string) error
	IsPasswordCorrect(user string, password string) bool
	HashAndSaveCookie(user string, cookie string, expirationDate time.Time) error
	IsCookieExpired(cookie string) bool
	GetUserViaCookie(cookie string) (string, error)
	ChangePassword(user string, newPassword string) error
	Logout(user string) error
	IsThereEnoughSpaceToAddVersion(user string, bytesToAdd int) error
	GetUsedSpaceInBytes(user string) (int, error)
	WipeDatabase()
}
