package versions

import (
	"ocelot/store/apps"
	"ocelot/store/users"
	"strconv"

	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"github.com/ocelot-cloud/shared/validation"
)

var (
	NotOwningThisVersionError = "you do not own this version"
	VersionDoesNotExistError  = "version does not exist"
	VersionAlreadyExist       = "version already exists"
	AppDoesNotExist           = "app does not exist"
)

type VersionService struct {
	VersionRepo VersionRepository
	UserService *users.UserServiceImpl
	AppRepo     apps.AppRepository // TODO !! not sure if needed
	AppService  *apps.AppServiceImpl
}

func (s *VersionService) DeleteVersionWithChecks(userId, versionId int) error {
	doesExist, errs := s.VersionRepo.DoesVersionIdExist(versionId)
	if errs != nil {
		return errs
	}
	if !doesExist {
		return u.Logger.NewError(VersionDoesNotExistError)
	}

	doesUserOwnVersion, err := s.VersionRepo.DoesUserOwnVersion(userId, versionId)
	if err != nil {
		return err
	}
	if !doesUserOwnVersion {
		return u.Logger.NewError(NotOwningThisVersionError)
	}

	err = s.VersionRepo.DeleteVersion(versionId)
	return err
}

func (s *VersionService) UploadVersion(userId int, versionUpload *store.VersionUploadDto) error {
	err := validation.ValidateStruct(versionUpload)
	if err != nil {
		return err
	}

	err = s.UserService.IsThereEnoughSpaceToAddVersion(userId, len(versionUpload.Content))
	if err != nil {
		return err
	}

	appId, err := strconv.Atoi(versionUpload.AppId)
	if err != nil {
		return u.Logger.NewError("could not convert to number")
	}

	if !s.AppRepo.DoesAppIdExist(appId) {
		return u.Logger.NewError(AppDoesNotExist)
	}

	isOwner, err := s.AppService.DoesUserOwnApp(userId, appId)
	if err != nil {
		return err
	}
	if !isOwner {
		return u.Logger.NewError(NotOwningThisVersionError)
	}

	app, err := s.AppRepo.GetAppById(appId)
	if err != nil {
		return err
	}

	err = validation.ValidateVersion(versionUpload.Content, app.Maintainer, app.Name)
	if err != nil {
		return err
	}

	doesVersionExist, err := s.VersionRepo.DoesVersionNameExist(appId, versionUpload.Version)
	if err != nil {
		return err
	}

	if doesVersionExist {
		return u.Logger.NewError(VersionAlreadyExist)
	}

	err = s.VersionRepo.CreateVersion(appId, versionUpload.Version, versionUpload.Content)
	if err != nil {
		return err
	}
	return nil
}

func (s *VersionService) ListVersions(appId int) ([]store.LeanVersionDto, error) {
	if !s.AppRepo.DoesAppIdExist(appId) {
		return nil, u.Logger.NewError(AppDoesNotExist)
	}
	versionsList, err := s.VersionRepo.ListVersionsOfApp(appId)
	if err != nil {
		return nil, err
	}
	return versionsList, nil
}

func (s *VersionService) GetVersionForDownload(versionId int) (*store.Version, error) {
	doesExist, err := s.VersionRepo.DoesVersionIdExist(versionId)
	if err != nil {
		return nil, err
	}
	if !doesExist {
		return nil, u.Logger.NewError(VersionDoesNotExistError)
	}
	versionInfo, err := s.VersionRepo.GetVersion(versionId)
	if err != nil {
		return nil, err
	}
	return versionInfo, nil
}
