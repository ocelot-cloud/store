package apps

import (
	"ocelot/store/users"

	u "github.com/ocelot-cloud/shared/utils"
)

var (
	AppNameReservedError  = "app name is reserved"
	AppAlreadyExistsError = "app already exists"
)

type AppServiceImpl struct {
	AppRepo  AppRepository
	UserRepo users.UserRepository
}

func (a *AppServiceImpl) DoesUserOwnApp(userId, appId int) (bool, error) {
	actualUserId, err := a.AppRepo.GetUserIdOfApp(appId)
	if err != nil {
		return false, err
	}
	return actualUserId == userId, nil
}

func (a *AppServiceImpl) CreateAppWithChecks(userId int, appName string) error {
	if appName == "ocelotcloud" {
		return u.Logger.NewError(AppNameReservedError)
	}

	doesExist, err := a.AppRepo.DoesAppExist(userId, appName)
	if err != nil {
		return err
	}
	if doesExist {
		return u.Logger.NewError(AppAlreadyExistsError)
	}

	err = a.AppRepo.CreateApp(userId, appName)
	if err != nil {
		return err
	}
	return nil
}
