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

// TODO !! should be user ID
func (s VersionService) DeleteVersionWithChecks(user string, versionId int) error {
	doesExist, errs := s.VersionRepo.DoesVersionExistTemp(versionId)
	if errs != nil {
		return errs
	}
	if !doesExist {
		return u.Logger.NewError(VersionDoesNotExistError)
	}

	// TODO !! should return error
	if !s.VersionRepo.DoesUserOwnVersion(user, versionId) {
		return u.Logger.NewError(NotOwningThisVersionError)
	}

	err := s.VersionRepo.DeleteVersion(versionId)
	return err
}
