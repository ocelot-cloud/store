package users

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"ocelot/store/tools"
	"time"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"golang.org/x/crypto/bcrypt"
)

// TODO !! add option to change email address; maybe make a field like "was email verified"?

// TODO !! simplify to CRUD operations, rest should be handle by a service
type UserRepository interface {
	// TODO !! keep functions
	// TODO CreateUser(form *store.RegistrationForm) error
	ValidateUserViaRegistrationCode(code string) error
	DoesUserExist(user string) bool
	DoesEmailExist(email string) bool
	DeleteUser(user string) error
	GetUserViaCookie(hashedCookieValue string) (*tools.User, error)
	ChangePassword(userId int, newPassword string) error
	Logout(user string) error

	// TODO !! replace functions
	CreateUserAndReturnRegistrationCode(form *store.RegistrationForm) (string, error) // TODo !! contains business logic
	IsPasswordCorrect(user string, password string) bool
	HashAndSaveCookie(user string, cookie string, expirationDate time.Time) error
	IsCookieExpired(cookie string) bool
	IsThereEnoughSpaceToAddVersion(user string, bytesToAdd int) error
	GetUsedSpaceInBytes(user string) (int, error)
	WipeDatabase()
	GetUserId(user string) (int, error)
	CreateAndValidateUser(form *store.RegistrationForm) error
}

type UserRepositoryImpl struct {
	DatabaseProvider *tools.DatabaseProviderImpl
	EmailVerifier    *tools.EmailVerifierImpl
	Config           *tools.Config
}

var NotEnoughSpacePrefix = "not enough space"

func (r *UserRepositoryImpl) IsThereEnoughSpaceToAddVersion(user string, bytesToAdd int) error {
	bytesUsed, err := r.GetUsedSpaceInBytes(user)
	if err != nil {
		u.Logger.Error("checking space failed", deepstack.ErrorField, err)
		return errors.New("checking space failed")
	}
	if bytesUsed+bytesToAdd > tools.MaxStorageSize {
		u.Logger.Info("user tried to upload version, but storage limit would be exceeded", tools.UserField, user)
		usedStorageInPercent := bytesUsed * 100 / tools.MaxStorageSize
		msg := fmt.Sprintf(NotEnoughSpacePrefix+", you can't store more then 10MiB of version content, currently used storage in bytes: %d/%d (%d percent)", bytesUsed, tools.MaxStorageSize, usedStorageInPercent)
		return errors.New(msg)
	}
	return nil
}

func (r *UserRepositoryImpl) IsPasswordCorrect(user string, password string) bool {
	var hashedPassword string
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT hashed_password FROM users WHERE user_name = $1", user).Scan(&hashedPassword)
	if err != nil {
		u.Logger.Error("Failed to fetch hashed password", deepstack.ErrorField, err)
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (r *UserRepositoryImpl) DoesUserExist(user string) bool {
	var exists bool
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_name = $1)", user).Scan(&exists)
	if err != nil {
		u.Logger.Error("Failed to check user existence", deepstack.ErrorField, err)
		return false
	}
	return exists
}

func (r *UserRepositoryImpl) CreateUserAndReturnRegistrationCode(form *store.RegistrationForm) (string, error) {
	var key string
	// TODO !! quite implicit logic. Maybe a better option to say in test mode, when we create a user, his account needs no code for validation
	if r.Config.CreateSampleData {
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

func (r *UserRepositoryImpl) ValidateUserViaRegistrationCode(code string) error {
	form, err := r.EmailVerifier.Load(code)
	if err != nil {
		return err
	}

	hashedPassword, err := u.SaltAndHash(form.Password)
	if err != nil {
		u.Logger.Error("Failed to hash password", deepstack.ErrorField, err)
		return fmt.Errorf("failed to hash password")
	}
	_, err = r.DatabaseProvider.GetDb().Exec("INSERT INTO users (user_name, email, hashed_password, used_space) VALUES ($1, $2, $3, $4)", form.User, form.Email, hashedPassword, 0)
	if err != nil {
		u.Logger.Error("Failed to create user", deepstack.ErrorField, err)
		return fmt.Errorf("failed to create user")
	}
	r.EmailVerifier.Delete(code)
	return nil
}

func (r *UserRepositoryImpl) DeleteUser(user string) error {
	if !r.DoesUserExist(user) {
		u.Logger.Info("User does not exist", deepstack.ErrorField, user)
		return fmt.Errorf("user does not exist")
	}

	_, err := r.DatabaseProvider.GetDb().Exec("DELETE FROM users WHERE user_name = $1", user)
	if err != nil {
		u.Logger.Error("Failed to delete user", deepstack.ErrorField, err)
		return fmt.Errorf("failed to delete user")
	}

	return nil
}

func (r *UserRepositoryImpl) HashAndSaveCookie(user string, cookie string, expirationDate time.Time) error {
	hashedCookieValue := u.GetSHA256Hash(cookie)

	_, err := r.DatabaseProvider.GetDb().Exec("UPDATE users SET hashed_cookie_value = $1, expiration_date = $2 WHERE user_name = $3", hashedCookieValue, expirationDate.Format(time.RFC3339), user)
	if err != nil {
		u.Logger.Error("Failed to hash and save cookie", deepstack.ErrorField, err)
		return fmt.Errorf("failed to hash and save cookie")
	}
	return nil
}

func (r *UserRepositoryImpl) IsCookieExpired(cookie string) bool {
	hashedCookieValue := u.GetSHA256Hash(cookie)

	var expirationDateStr string
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT expiration_date FROM users WHERE hashed_cookie_value = $1", hashedCookieValue).Scan(&expirationDateStr)
	if err != nil {
		u.Logger.Error("Failed to fetch expiration date", deepstack.ErrorField, err)
		return true
	} else if expirationDateStr == "" {
		return true
	}

	expirationDate, err := time.Parse(time.RFC3339, expirationDateStr)
	if err != nil {
		u.Logger.Error("Failed to parse expiration date", deepstack.ErrorField, err)
		return true
	}

	return time.Now().UTC().After(expirationDate)
}

func (r *UserRepositoryImpl) GetUserViaCookie(hashedCookieValue string) (*tools.User, error) {
	var user tools.User
	err := r.DatabaseProvider.GetDb().QueryRow(
		`SELECT user_id, user_name, email, hashed_password, hashed_cookie_value, expiration_date, used_space 
		 FROM users WHERE hashed_cookie_value = $1`,
		hashedCookieValue,
	).Scan(
		&user.UserId,
		&user.UserName,
		&user.Email,
		&user.HashedPassword,
		&user.HashedCookieValue,
		&user.ExpirationDate,
		&user.UsedSpace,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			u.Logger.Info("Cookie not found")
			return nil, fmt.Errorf("cookie not found")
		}
		u.Logger.Error("Failed to fetch user", deepstack.ErrorField, err)
		return nil, fmt.Errorf("failed to fetch user")
	}

	return &user, nil
}

func (r *UserRepositoryImpl) ChangePassword(userId int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		u.Logger.Error("Failed to hash password", deepstack.ErrorField, err)
		return fmt.Errorf("failed to hash password")
	}

	_, err = r.DatabaseProvider.GetDb().Exec("UPDATE users SET hashed_password = $1 WHERE user_id = $2", hashedPassword, userId)
	if err != nil {
		u.Logger.Error("Failed to change password", deepstack.ErrorField, err)
		return fmt.Errorf("failed to change password")
	}

	return nil
}

func (r *UserRepositoryImpl) WipeDatabase() {
	_, err := r.DatabaseProvider.GetDb().Exec("DELETE FROM users WHERE user_name != 'sample'")
	if err != nil {
		u.Logger.Error("Failed to wipe database", deepstack.ErrorField, err)
	}
	r.EmailVerifier.Clear()
}

func (r *UserRepositoryImpl) GetUsedSpaceInBytes(user string) (int, error) {
	var usedSpace int
	err := r.DatabaseProvider.GetDb().QueryRow(`SELECT used_space FROM users WHERE user_name = $1`, user).Scan(&usedSpace)
	if err != nil {
		u.Logger.Error("Failed to get used space", deepstack.ErrorField, err)
		return 0, fmt.Errorf("failed to get used space")
	}
	return usedSpace, nil
}

func (r *UserRepositoryImpl) Logout(user string) error {
	_, err := r.DatabaseProvider.GetDb().Exec("UPDATE users SET hashed_cookie_value = $1, expiration_date = $2 WHERE user_name = $3", nil, nil, user)
	if err != nil {
		u.Logger.Error("failed to logout", deepstack.ErrorField, err)
		return errors.New("failed to logout")
	}
	return nil
}

func (r *UserRepositoryImpl) CreateAndValidateUser(form *store.RegistrationForm) error {
	code, err := r.CreateUserAndReturnRegistrationCode(form)
	if err != nil {
		return err
	}
	err = r.ValidateUserViaRegistrationCode(code)
	return err
}

func (r *UserRepositoryImpl) DoesEmailExist(email string) bool {
	var exists bool
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		u.Logger.Error("Failed to check email existence", deepstack.ErrorField, err)
		return false
	}
	return exists
}

func (r *UserRepositoryImpl) GetUserId(user string) (int, error) {
	var userID int
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT user_id FROM users WHERE user_name = $1", user).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("user not found: %w", err)
	}
	return userID, nil
}
