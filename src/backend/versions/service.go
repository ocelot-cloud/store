package versions

import (
	u "github.com/ocelot-cloud/shared/utils"
)

var (
	NotOwningThisVersionError = "you do not own this version"
	VersionDoesNotExistError  = "version does not exist"
)

type VersionService struct {
	VersionRepo VersionRepository
}

func (s VersionService) DeleteVersionWithChecks(userId, versionId int) error {
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
