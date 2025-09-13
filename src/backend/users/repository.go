package users

import (
	"database/sql"
	"errors"
	"fmt"
	"ocelot/store/tools"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"golang.org/x/crypto/bcrypt"
)

var NotEnoughSpacePrefix = "not enough space"

type UserRepository interface {
	CreateUser(form *store.RegistrationForm) error
	DoesUserExist(user string) (bool, error)
	DoesEmailExist(email string) (bool, error)
	DeleteUser(user string) error
	GetUserViaCookie(hashedCookieValue string) (*tools.User, error)
	ChangePassword(userId int, newPassword string) error
	Logout(user string) error
	GetUserByName(user string) (*tools.User, error)
	UpdateUser(*tools.User) error
	GetUserById(userId int) (*tools.User, error)
	WipeUsers()
}

type UserRepositoryImpl struct {
	DatabaseProvider *tools.DatabaseProviderImpl
}

func (r *UserRepositoryImpl) GetUserById(userId int) (*tools.User, error) {
	var user tools.User
	err := r.DatabaseProvider.GetDb().QueryRow(
		`SELECT 
			user_id, 
			user_name, 
			email, 
			hashed_password, 
			hashed_cookie_value,
			expiration_date, 
			used_space_in_bytes
		 FROM users 
		 WHERE user_id = $1`,
		userId,
	).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.HashedPassword,
		&user.HashedCookieValue,
		&user.ExpirationDate,
		&user.UsedSpaceInBytes,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) UpdateUser(user *tools.User) error {
	_, err := r.DatabaseProvider.GetDb().Exec(
		`UPDATE users 
		 SET user_name = $1,
		     email = $2,
		     hashed_password = $3,
		     hashed_cookie_value = $4,
		     expiration_date = $5,
		     used_space_in_bytes = $6
		 WHERE user_id = $7`,
		user.Name,
		user.Email,
		user.HashedPassword,
		user.HashedCookieValue,
		user.ExpirationDate,
		user.UsedSpaceInBytes,
		user.Id,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *UserRepositoryImpl) GetUserByName(userName string) (*tools.User, error) {
	var user tools.User
	err := r.DatabaseProvider.GetDb().QueryRow(
		`SELECT 
			user_id, 
			user_name, 
			email, 
			hashed_password, 
			hashed_cookie_value,
			expiration_date, 
			used_space_in_bytes
		 FROM users 
		 WHERE user_name = $1`,
		userName,
	).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.HashedPassword,
		&user.HashedCookieValue,
		&user.ExpirationDate,
		&user.UsedSpaceInBytes,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New(UserDoesNotExistError)
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) DoesUserExist(user string) (bool, error) {
	var exists bool
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_name = $1)", user).Scan(&exists)
	if err != nil {
		return false, u.Logger.NewError(err.Error())
	}
	return exists, nil
}

func (r *UserRepositoryImpl) CreateUser(form *store.RegistrationForm) error {
	// TODO !! can this operation ever fail? if not, remove the error returned
	hashedPassword, err := u.SaltAndHash(form.Password)
	if err != nil {
		u.Logger.Error("Failed to hash password", deepstack.ErrorField, err)
		return fmt.Errorf("failed to hash password")
	}
	_, err = r.DatabaseProvider.GetDb().Exec("INSERT INTO users (user_name, email, hashed_password, used_space_in_bytes) VALUES ($1, $2, $3, $4)", form.User, form.Email, hashedPassword, 0)
	if err != nil {
		u.Logger.Error("Failed to create user", deepstack.ErrorField, err)
		return fmt.Errorf("failed to create user")
	}
	return nil
}

func (r *UserRepositoryImpl) DeleteUser(user string) error {
	_, err := r.DatabaseProvider.GetDb().Exec("DELETE FROM users WHERE user_name = $1", user)
	if err != nil {
		u.Logger.Error("Failed to delete user", deepstack.ErrorField, err)
		return fmt.Errorf("failed to delete user")
	}
	return nil
}

func (r *UserRepositoryImpl) GetUserViaCookie(hashedCookieValue string) (*tools.User, error) {
	var user tools.User
	err := r.DatabaseProvider.GetDb().QueryRow(
		`SELECT user_id, user_name, email, hashed_password, hashed_cookie_value, expiration_date, used_space_in_bytes 
		 FROM users WHERE hashed_cookie_value = $1`,
		hashedCookieValue,
	).Scan(
		&user.Id,
		&user.Name,
		&user.Email,
		&user.HashedPassword,
		&user.HashedCookieValue,
		&user.ExpirationDate,
		&user.UsedSpaceInBytes,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, u.Logger.NewError("cookie not found")
	}
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
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

func (r *UserRepositoryImpl) WipeUsers() {
	_, err := r.DatabaseProvider.GetDb().Exec("DELETE FROM users")
	if err != nil {
		u.Logger.Error("Failed to wipe database", deepstack.ErrorField, err)
	}
}

func (r *UserRepositoryImpl) Logout(user string) error {
	_, err := r.DatabaseProvider.GetDb().Exec("UPDATE users SET hashed_cookie_value = $1, expiration_date = $2 WHERE user_name = $3", nil, nil, user)
	if err != nil {
		u.Logger.Error("failed to logout", deepstack.ErrorField, err)
		return errors.New("failed to logout")
	}
	return nil
}

func (r *UserRepositoryImpl) DoesEmailExist(email string) (bool, error) {
	var exists bool
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return false, u.Logger.NewError(err.Error())
	}
	return exists, nil
}
