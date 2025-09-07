package versions

import (
	"fmt"
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"
	"strconv"
	"time"

	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
)

type VersionRepository interface {
	IsVersionOwner(user string, versionId int) bool
	CreateVersion(appId int, version string, data []byte) error
	GetVersionId(appId int, version string) (int, error)
	DeleteVersion(versionId int) error
	GetVersionList(appId int) ([]store.Version, error)
	DoesVersionExist(versionId int) bool
	GetVersionContent(versionId int) ([]byte, error)
	GetAppIdByVersionId(versionId int) (int, error)
	GetFullVersionInfo(versionId int) (*store.FullVersionInfo, error)
}

type VersionRepositoryImpl struct {
	DatabaseProvider *tools.DatabaseProviderImpl
	UserRepo         users.UserRepository
	AppRepo          apps.AppRepository
}

func (r *VersionRepositoryImpl) GetFullVersionInfo(versionId int) (*store.FullVersionInfo, error) {
	var fullVersionInfo store.FullVersionInfo
	err := r.DatabaseProvider.GetDb().QueryRow(`
		SELECT users.user_name, apps.app_name, versions.version_name, versions.data, versions.version_id, versions.creation_timestamp
		FROM versions
		JOIN apps ON versions.app_id = apps.app_id
		JOIN users ON apps.user_id = users.user_id
		WHERE versions.version_id = $1
	`, versionId).Scan(&fullVersionInfo.Maintainer, &fullVersionInfo.AppName, &fullVersionInfo.VersionName, &fullVersionInfo.Content, &fullVersionInfo.Id, &fullVersionInfo.VersionCreationTimestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to get full version info: %w", err)
	}
	return &fullVersionInfo, nil
}

func (r *VersionRepositoryImpl) GetAppIdByVersionId(versionId int) (int, error) {
	var appId int
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT app_id FROM versions WHERE version_id = $1", versionId).Scan(&appId)
	if err != nil {
		u.Logger.Error("Failed to get app ID by version ID", tools.VersionIdField, versionId, deepstack.ErrorField, err)
		return -1, fmt.Errorf("failed to get app ID by version ID")
	}
	return appId, nil
}

func (r *VersionRepositoryImpl) IsVersionOwner(user string, versionId int) bool {
	userId, err := r.UserRepo.GetUserId(user)
	if err != nil {
		u.Logger.Info("Failed to get user ID", tools.UserField, deepstack.ErrorField, err)
		return false
	}

	var ownerId int
	err = r.DatabaseProvider.GetDb().QueryRow(`
		SELECT apps.user_id 
		FROM versions
		JOIN apps ON versions.app_id = apps.app_id
		WHERE versions.version_id = $1`, versionId).Scan(&ownerId)
	if err != nil {
		u.Logger.Error("Failed to get version owner ID", deepstack.ErrorField, err)
		return false
	}

	return userId == ownerId
}

func (r *VersionRepositoryImpl) GetVersionContent(versionId int) ([]byte, error) {
	var data []byte
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT data FROM versions WHERE version_id = $1", versionId).Scan(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *VersionRepositoryImpl) CreateVersion(appId int, version string, data []byte) error {
	userId, err := r.AppRepo.GetUserIdOfApp(appId)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err = r.DatabaseProvider.GetDb().Exec("INSERT INTO versions (app_id, version_name, creation_timestamp, data) VALUES ($1, $2, $3, $4)", appId, version, now, data)
	if err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}

	dataSize := len(data)
	_, err = r.DatabaseProvider.GetDb().Exec("UPDATE users SET used_space = used_space + $1 WHERE user_id = $2", dataSize, userId)
	if err != nil {
		return fmt.Errorf("failed to update user space: %w", err)
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
		return fmt.Errorf("failed to delete version: %w", err)
	}

	_, err = r.DatabaseProvider.GetDb().Exec("UPDATE users SET used_space = used_space - $1 WHERE user_id = $2", dataSize, userId)
	if err != nil {
		return fmt.Errorf("failed to update user space: %w", err)
	}

	return nil
}

func (r *VersionRepositoryImpl) getAppIdOfVersion(versionId int) (int, error) {
	var appId int
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT app_id FROM versions WHERE version_id = $1", versionId).Scan(&appId)
	if err != nil {
		return -1, fmt.Errorf("failed to get app ID: %w", err)
	}
	return appId, nil
}

func (r *VersionRepositoryImpl) getBlobSize(versionId int) (int64, error) {
	var dataSize int64
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT LENGTH(data) FROM versions WHERE version_id = $1", versionId).Scan(&dataSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get BLOB size: %w", err)
	}
	return dataSize, nil
}

func (r *VersionRepositoryImpl) GetVersionList(appId int) ([]store.Version, error) {
	var exists bool
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT EXISTS(SELECT 1 FROM apps WHERE app_id = $1)", appId).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check app existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("app with id %d does not exist", appId)
	}

	rows, err := r.DatabaseProvider.GetDb().Query("SELECT version_name, version_id, creation_timestamp FROM versions WHERE app_id = $1 ORDER BY creation_timestamp DESC", appId)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}
	defer u.Close(rows)

	var versions []store.Version
	for rows.Next() {
		var version string
		var id int
		var creationTimestamp time.Time
		if err := rows.Scan(&version, &id, &creationTimestamp); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		creationTimestamp = creationTimestamp.UTC()
		versions = append(versions, store.Version{
			Name:              version,
			Id:                strconv.Itoa(id),
			CreationTimestamp: creationTimestamp,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return versions, nil
}

func (r *VersionRepositoryImpl) DoesVersionExist(versionId int) bool {
	var exists bool
	err := r.DatabaseProvider.GetDb().QueryRow(`SELECT EXISTS(SELECT 1 FROM versions WHERE version_id = $1)`, versionId).Scan(&exists)
	if err != nil {
		u.Logger.Debug("error checking if version exists")
		return false
	}
	return exists
}

func (r *VersionRepositoryImpl) GetVersionId(appId int, version string) (int, error) {
	var versionId int
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT version_id FROM versions WHERE app_id = $1 AND version_name = $2", appId, version).Scan(&versionId)
	if err != nil {
		return -1, fmt.Errorf("version not found: %w", err)
	}
	return versionId, nil
}
