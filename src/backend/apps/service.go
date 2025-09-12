package apps

type AppServiceImpl struct {
	AppRepo AppRepository
}

func (a *AppServiceImpl) DoesUserOwnApp(userId, appId int) (bool, error) {
	actualUserId, err := a.AppRepo.GetUserIdOfApp(appId)
	if err != nil {
		return false, err
	}
	return actualUserId == userId, nil
}
