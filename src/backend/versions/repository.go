package versions

import (
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"
	"strconv"
	"time"

	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
)

type VersionRepository interface {
	DoesVersionIdExist(versionId int) (bool, error)
	DoesVersionNameExist(appId int, version string) (bool, error)
	DoesUserOwnVersion(userId, versionId int) (bool, error)
	CreateVersion(appId int, version string, data []byte) error
	DeleteVersion(versionId int) error
	ListVersionsOfApp(appId int) ([]store.LeanVersionDto, error)
	GetVersion(versionId int) (*store.Version, error)
}

type VersionRepositoryImpl struct {
	DatabaseProvider *tools.DatabaseProviderImpl
	UserRepo         users.UserRepository
	AppRepo          apps.AppRepository
}

func (r *VersionRepositoryImpl) GetVersion(versionId int) (*store.Version, error) {
	var version store.Version
	err := r.DatabaseProvider.GetDb().QueryRow(`
		SELECT users.user_name, apps.app_name, versions.version_name, versions.data, versions.version_id, versions.creation_timestamp
		FROM versions
		JOIN apps ON versions.app_id = apps.app_id
		JOIN users ON apps.user_id = users.user_id
		WHERE versions.version_id = $1
	`, versionId).Scan(&version.Maintainer, &version.AppName, &version.VersionName, &version.Content, &version.Id, &version.VersionCreationTimestamp)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return &version, nil
}

func (r *VersionRepositoryImpl) DoesUserOwnVersion(userId, versionId int) (bool, error) {
	var ownerId int
	err := r.DatabaseProvider.GetDb().QueryRow(`
		SELECT apps.user_id 
		FROM versions
		JOIN apps ON versions.app_id = apps.app_id
		WHERE versions.version_id = $1`, versionId).Scan(&ownerId)
	if err != nil {
		return false, u.Logger.NewError(err.Error())
	}

	return userId == ownerId, nil
}
func (r *VersionRepositoryImpl) CreateVersion(appId int, version string, data []byte) error {
	userId, err := r.AppRepo.GetUserIdOfApp(appId)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err = r.DatabaseProvider.GetDb().Exec("INSERT INTO versions (app_id, version_name, creation_timestamp, data) VALUES ($1, $2, $3, $4)", appId, version, now, data)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}

	dataSize := len(data)
	_, err = r.DatabaseProvider.GetDb().Exec("UPDATE users SET used_space_in_bytes = used_space_in_bytes + $1 WHERE user_id = $2", dataSize, userId)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}

	return nil
}

func (r *VersionRepositoryImpl) DeleteVersion(versionId int) error {
	dataSize, err := r.getBlobSize(versionId)
	if err != nil {
		return err
	}
	appId, err := r.getAppIdOfVersion(versionId)
	if err != nil {
		return err
	}
	userId, err := r.AppRepo.GetUserIdOfApp(appId)
	if err != nil {
		return err
	}

	_, err = r.DatabaseProvider.GetDb().Exec("DELETE FROM versions WHERE version_id = $1", versionId)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}

	_, err = r.DatabaseProvider.GetDb().Exec("UPDATE users SET used_space_in_bytes = used_space_in_bytes - $1 WHERE user_id = $2", dataSize, userId)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}

	return nil
}

func (r *VersionRepositoryImpl) getAppIdOfVersion(versionId int) (int, error) {
	var appId int
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT app_id FROM versions WHERE version_id = $1", versionId).Scan(&appId)
	if err != nil {
		return -1, u.Logger.NewError(err.Error())
	}
	return appId, nil
}

func (r *VersionRepositoryImpl) getBlobSize(versionId int) (int64, error) {
	var dataSize int64
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT LENGTH(data) FROM versions WHERE version_id = $1", versionId).Scan(&dataSize)
	if err != nil {
		return 0, u.Logger.NewError(err.Error())
	}
	return dataSize, nil
}

func (r *VersionRepositoryImpl) ListVersionsOfApp(appId int) ([]store.LeanVersionDto, error) {
	// TODO !! this kind of check is business logic
	var exists bool
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT EXISTS(SELECT 1 FROM apps WHERE app_id = $1)", appId).Scan(&exists)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	if !exists {
		return nil, u.Logger.NewError(AppDoesNotExist)
	}

	rows, err := r.DatabaseProvider.GetDb().Query("SELECT version_name, version_id, creation_timestamp FROM versions WHERE app_id = $1", appId)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(rows)

	var versions []store.LeanVersionDto
	for rows.Next() {
		var version string
		var id int
		var creationTimestamp time.Time
		if err = rows.Scan(&version, &id, &creationTimestamp); err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		creationTimestamp = creationTimestamp.UTC()
		versions = append(versions, store.LeanVersionDto{
			Id:                strconv.Itoa(id),
			Name:              version,
			CreationTimestamp: creationTimestamp,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	return versions, nil
}

func (r *VersionRepositoryImpl) DoesVersionIdExist(versionId int) (bool, error) {
	var exists bool
	err := r.DatabaseProvider.GetDb().QueryRow(`SELECT EXISTS(SELECT 1 FROM versions WHERE version_id = $1)`, versionId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *VersionRepositoryImpl) DoesVersionNameExist(appId int, version string) (bool, error) {
	var exists bool
	err := r.DatabaseProvider.GetDb().QueryRow(
		`SELECT EXISTS(SELECT 1 FROM versions WHERE app_id = $1 AND version_name = $2)`,
		appId, version,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
