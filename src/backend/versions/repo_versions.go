package versions

import (
	"fmt"
	"github.com/ocelot-cloud/shared/store"
	"github.com/ocelot-cloud/shared/utils"
	"ocelot/store/apps"
	"ocelot/store/tools"
	"strconv"
	"time"
)

var VersionRepo VersionRepository = &VersionRepositoryImpl{}

func (u *VersionRepositoryImpl) GetFullVersionInfo(versionId int) (*store.FullVersionInfo, error) {
	var fullVersionInfo store.FullVersionInfo
	err := tools.Db.QueryRow(`
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

func (u *VersionRepositoryImpl) GetAppIdByVersionId(versionId int) (int, error) {
	var appId int
	err := tools.Db.QueryRow("SELECT app_id FROM versions WHERE version_id = $1", versionId).Scan(&appId)
	if err != nil {
		tools.Logger.Error("Failed to get app ID by version ID %d: %v", versionId, err)
		return -1, fmt.Errorf("failed to get app ID by version ID")
	}
	return appId, nil
}

func (u *VersionRepositoryImpl) IsVersionOwner(user string, versionId int) bool {
	userId, err := tools.GetUserId(user)
	if err != nil {
		tools.Logger.Info("Failed to get user ID: %v", err)
		return false
	}

	var ownerId int
	err = tools.Db.QueryRow(`
		SELECT apps.user_id 
		FROM versions
		JOIN apps ON versions.app_id = apps.app_id
		WHERE versions.version_id = $1`, versionId).Scan(&ownerId)
	if err != nil {
		tools.Logger.Error("Failed to get version owner ID: %v", err)
		return false
	}

	return userId == ownerId
}

func (u *VersionRepositoryImpl) GetVersionContent(versionId int) ([]byte, error) {
	var data []byte
	err := tools.Db.QueryRow("SELECT data FROM versions WHERE version_id = $1", versionId).Scan(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (u *VersionRepositoryImpl) CreateVersion(appId int, version string, data []byte) error {
	userId, err := apps.GetUserIdOfApp(appId)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err = tools.Db.Exec("INSERT INTO versions (app_id, version_name, creation_timestamp, data) VALUES ($1, $2, $3, $4)", appId, version, now, data)
	if err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}

	dataSize := len(data)
	_, err = tools.Db.Exec("UPDATE users SET used_space = used_space + $1 WHERE user_id = $2", dataSize, userId)
	if err != nil {
		return fmt.Errorf("failed to update user space: %w", err)
	}

	return nil
}

func (u *VersionRepositoryImpl) DeleteVersion(versionId int) error {
	dataSize, err := getBlobSize(versionId)
	if err != nil {
		return err
	}
	appId, err := getAppIdOfVersion(versionId)
	if err != nil {
		return err
	}
	userId, err := apps.GetUserIdOfApp(appId)
	if err != nil {
		return err
	}

	_, err = tools.Db.Exec("DELETE FROM versions WHERE version_id = $1", versionId)
	if err != nil {
		return fmt.Errorf("failed to delete version: %w", err)
	}

	_, err = tools.Db.Exec("UPDATE users SET used_space = used_space - $1 WHERE user_id = $2", dataSize, userId)
	if err != nil {
		return fmt.Errorf("failed to update user space: %w", err)
	}

	return nil
}

func getAppIdOfVersion(versionId int) (int, error) {
	var appId int
	err := tools.Db.QueryRow("SELECT app_id FROM versions WHERE version_id = $1", versionId).Scan(&appId)
	if err != nil {
		return -1, fmt.Errorf("failed to get app ID: %w", err)
	}
	return appId, nil
}

func getBlobSize(versionId int) (int64, error) {
	var dataSize int64
	err := tools.Db.QueryRow("SELECT LENGTH(data) FROM versions WHERE version_id = $1", versionId).Scan(&dataSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get BLOB size: %w", err)
	}
	return dataSize, nil
}

func (u *VersionRepositoryImpl) GetVersionList(appId int) ([]store.Version, error) {
	var exists bool
	err := tools.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM apps WHERE app_id = $1)", appId).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check app existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("app with id %d does not exist", appId)
	}

	rows, err := tools.Db.Query("SELECT version_name, version_id, creation_timestamp FROM versions WHERE app_id = $1 ORDER BY creation_timestamp DESC", appId)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}
	defer utils.Close(rows)

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

func (u *VersionRepositoryImpl) DoesVersionExist(versionId int) bool {
	var exists bool
	err := tools.Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM versions WHERE version_id = $1)`, versionId).Scan(&exists)
	if err != nil {
		tools.Logger.Debug("error checking if version exists")
		return false
	}
	return exists
}

func (u *VersionRepositoryImpl) GetVersionId(appId int, version string) (int, error) {
	var versionId int
	err := tools.Db.QueryRow("SELECT version_id FROM versions WHERE app_id = $1 AND version_name = $2", appId, version).Scan(&versionId)
	if err != nil {
		return -1, fmt.Errorf("version not found: %w", err)
	}
	return versionId, nil
}

type VersionRepositoryImpl struct{}

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
