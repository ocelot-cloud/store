package apps

import (
	"database/sql"
	"fmt"
	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"ocelot/store/tools"
	"ocelot/store/users"
	"strconv"
)

// TODO !! rename repo_apps, repo_users and repo_versions simply to repository each

var (
	// TODO !! global var
	AppRepo AppRepository = &AppRepositoryImpl{}
)

func (r *AppRepositoryImpl) IsAppOwner(user string, appId int) bool {
	userId, err := tools.GetUserId(user)
	if err != nil {
		u.Logger.Info("Failed to get user ID", deepstack.ErrorField, err)
		return false
	}

	var ownerId int
	err = tools.Db.QueryRow("SELECT user_id FROM apps WHERE app_id = $1", appId).Scan(&ownerId)
	if err != nil {
		u.Logger.Error("Failed to get app owner ID", deepstack.ErrorField, err)
		return false
	}

	return userId == ownerId
}

func (r *AppRepositoryImpl) CreateApp(user string, app string) error {
	if !users.UserRepo.DoesUserExist(user) {
		u.Logger.Info("User does not exist", tools.UserField, user)
		return fmt.Errorf("user does not exist")
	}

	userID, err := tools.GetUserId(user)
	if err != nil {
		return err
	}
	_, err = tools.Db.Exec(`INSERT INTO apps (user_id, app_name) VALUES ($1, $2)`, userID, app)
	if err != nil {
		u.Logger.Error("Failed to create app", deepstack.ErrorField, err)
		return fmt.Errorf("failed to create app")
	}
	return nil
}

func (r *AppRepositoryImpl) DoesAppExist(appId int) bool {
	var exists bool
	err := tools.Db.QueryRow("SELECT EXISTS(SELECT 1 FROM apps WHERE app_id = $1)", appId).Scan(&exists)
	if err != nil {
		u.Logger.Error("Failed to check app existence for app", tools.AppIdField, appId, deepstack.ErrorField, err)
		return false
	}
	return exists
}

func (r *AppRepositoryImpl) DeleteApp(appId int) error {
	userId, err := GetUserIdOfApp(appId)
	if err != nil {
		return err
	}

	totalDataSize, err := r.sumBlobSizes(appId)
	if err != nil {
		return err
	}

	_, err = tools.Db.Exec(`DELETE FROM apps WHERE app_id = $1`, appId)
	if err != nil {
		u.Logger.Error("Failed to delete app", deepstack.ErrorField, err)
		return fmt.Errorf("failed to delete app")
	}

	_, err = tools.Db.Exec("UPDATE users SET used_space = used_space - $1 WHERE user_id = $2", totalDataSize, userId)
	if err != nil {
		return fmt.Errorf("failed to update user space: %w", err)
	}

	return nil
}

func GetUserIdOfApp(appId int) (int, error) {
	var userId int
	err := tools.Db.QueryRow(`SELECT user_id FROM apps WHERE app_id = $1`, appId).Scan(&userId)
	if err != nil {
		u.Logger.Error("Failed to get user ID of app", tools.AppIdField, appId, deepstack.ErrorField, err)
		return -1, fmt.Errorf("failed to get user ID of app")
	}
	return userId, nil
}

func (r *AppRepositoryImpl) sumBlobSizes(appID int) (int64, error) {
	var totalSize sql.NullInt64
	err := tools.Db.QueryRow("SELECT SUM(LENGTH(data)) FROM versions WHERE app_id = $1", appID).Scan(&totalSize)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total BLOB size: %w", err)
	}

	if !totalSize.Valid {
		return 0, nil
	}

	return totalSize.Int64, nil
}

func (r *AppRepositoryImpl) SearchForApps(request store.AppSearchRequest) ([]store.AppWithLatestVersion, error) {
	var apps []store.AppWithLatestVersion
	query := `
		SELECT u.user_name, a.app_id, a.app_name, v.version_id, v.version_name
		FROM users u
		JOIN apps a ON u.user_id = a.user_id
		JOIN LATERAL (
			SELECT version_id, version_name
			FROM versions
			WHERE app_id = a.app_id
			ORDER BY creation_timestamp DESC
			LIMIT 1
		) v ON true
		WHERE (u.user_name LIKE $1 OR a.app_name LIKE $2)
	`

	if !request.ShowUnofficialApps {
		query += " AND u.user_name = 'ocelotcloud'"
	}
	query += " LIMIT 100"

	rows, err := tools.Db.Query(query, "%"+request.SearchTerm+"%", "%"+request.SearchTerm+"%")
	if err != nil {
		u.Logger.Error("Failed to find apps", deepstack.ErrorField, err)
		return nil, fmt.Errorf("failed to find apps")
	}
	defer u.Close(rows)

	for rows.Next() {
		var maintainer, appName, versionName string
		var appId, versionId int
		err := rows.Scan(&maintainer, &appId, &appName, &versionId, &versionName)
		if err != nil {
			u.Logger.Error("Error scanning app row", deepstack.ErrorField, err)
			continue
		}
		apps = append(apps, store.AppWithLatestVersion{
			Maintainer:        maintainer,
			AppId:             strconv.Itoa(appId),
			AppName:           appName,
			LatestVersionId:   strconv.Itoa(versionId),
			LatestVersionName: versionName,
		})
	}
	err = rows.Err()
	if err != nil {
		u.Logger.Error("Error iterating over rows", deepstack.ErrorField, err)
		return nil, fmt.Errorf("error iterating over rows")
	}
	return apps, nil
}

func (r *AppRepositoryImpl) GetAppId(user, app string) (int, error) {
	userID, err := tools.GetUserId(user)
	if err != nil {
		return -1, err
	}

	appID, err := tools.GetAppId(userID, app)
	if err != nil {
		return -1, err
	}
	return appID, nil
}

func (r *AppRepositoryImpl) GetAppList(user string) ([]store.App, error) {
	userID, err := tools.GetUserId(user)
	if err != nil {
		return nil, err
	}

	rows, err := tools.Db.Query("SELECT app_name, app_id FROM apps WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apps: %w", err)
	}
	defer u.Close(rows)

	var apps []store.App
	for rows.Next() {
		var app string
		var appId int
		if err = rows.Scan(&app, &appId); err != nil {
			return nil, fmt.Errorf("failed to scan app: %w", err)
		}
		apps = append(apps, store.App{Maintainer: user, Name: app, Id: strconv.Itoa(appId)})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return apps, nil
}

func (r *AppRepositoryImpl) GetAppName(appId int) (string, error) {
	var appName string
	err := tools.Db.QueryRow("SELECT app_name FROM apps WHERE app_id = $1", appId).Scan(&appName)
	if err != nil {
		return "", fmt.Errorf("failed to get app name: %w", err)
	}
	return appName, nil
}

func (r *AppRepositoryImpl) GetMaintainerName(appId int) (string, error) {
	var maintainer string
	err := tools.Db.QueryRow(`
		SELECT u.user_name
		FROM users u
		JOIN apps a ON u.user_id = a.user_id
		WHERE a.app_id = $1
	`, appId).Scan(&maintainer)
	if err != nil {
		return "", fmt.Errorf("failed to get maintainer name: %w", err)
	}
	return maintainer, nil
}

type AppRepositoryImpl struct{}

type AppRepository interface {
	IsAppOwner(user string, appId int) bool
	DoesAppExist(appId int) bool
	CreateApp(user, app string) error
	DeleteApp(appId int) error
	SearchForApps(searchRequest store.AppSearchRequest) ([]store.AppWithLatestVersion, error)
	GetAppId(user, app string) (int, error)
	GetAppName(appId int) (string, error)
	GetAppList(user string) ([]store.App, error)
	GetMaintainerName(appId int) (string, error)
}
