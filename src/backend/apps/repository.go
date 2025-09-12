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

// TODO !! never pass username if you could pass user ID instead
type AppRepository interface {
	// TODO !! keep functions
	DoesUserOwnApp(userId, appId int) bool
	DoesAppExist(appId int) bool
	CreateApp(userId int, app string) error
	DeleteApp(appId int) error
	GetAppList(userId int) ([]store.AppDto, error)
	SearchForApps(searchRequest store.AppSearchRequest) ([]store.AppWithLatestVersion, error) // TODO !! not sure whether it makes sense to maybe improve my search function, like explicitly say have a field for maintainer and app you can search for; if empty, its ignored
	GetAppById(appId int) (store.AppDto, error)

	// TODO !! remove functions
	GetAppName(appId int) (string, error)
	GetMaintainerName(appId int) (string, error)
	// TODO !! duplication, only give ID? or maybe pass the user struct
	GetUserIdOfApp(appId int) (int, error)
	GetAppId(userID int, app string) (int, error) // TODO !! looks like sth I dont need?
}

type AppRepositoryImpl struct {
	DatabaseProvider *tools.DatabaseProviderImpl
	UserRepo         users.UserRepository
}

// TODO !! this is an AppDto. Within my application the ID should remain an integer, so I need store.App and AppDto
func (r *AppRepositoryImpl) GetAppById(appId int) (store.AppDto, error) {
	var app store.AppDto // TODO !! dont use DTO
	err := r.DatabaseProvider.GetDb().QueryRow(
		`SELECT u.user_name, a.app_name, a.app_id
		 FROM apps a
		 JOIN users u ON a.user_id = u.user_id
		 WHERE a.app_id = $1`,
		appId,
	).Scan(&app.Maintainer, &app.Name, &app.Id)
	if err != nil {
		// TODO !! dont use DTO
		return store.AppDto{}, fmt.Errorf("failed to get app by id: %w", err)
	}
	app.Id = strconv.Itoa(appId) // TODO !! remove this, should be an integer
	return app, nil
}

func (r *AppRepositoryImpl) DoesUserOwnApp(userId, appId int) bool {
	var ownerId int
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT user_id FROM apps WHERE app_id = $1", appId).Scan(&ownerId)
	if err != nil {
		u.Logger.Error("Failed to get app owner ID", deepstack.ErrorField, err)
		return false
	}

	return userId == ownerId
}

func (r *AppRepositoryImpl) CreateApp(userId int, app string) error {
	_, err := r.DatabaseProvider.GetDb().Exec(`INSERT INTO apps (user_id, app_name) VALUES ($1, $2)`, userId, app)
	if err != nil {
		u.Logger.Error("Failed to create app", deepstack.ErrorField, err)
		return fmt.Errorf("failed to create app")
	}
	return nil
}

func (r *AppRepositoryImpl) DoesAppExist(appId int) bool {
	var exists bool
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT EXISTS(SELECT 1 FROM apps WHERE app_id = $1)", appId).Scan(&exists)
	if err != nil {
		u.Logger.Error("Failed to check app existence for app", tools.AppIdField, appId, deepstack.ErrorField, err)
		return false
	}
	return exists
}

func (r *AppRepositoryImpl) DeleteApp(appId int) error {
	userId, err := r.GetUserIdOfApp(appId)
	if err != nil {
		return err
	}

	totalDataSize, err := r.sumBlobSizes(appId)
	if err != nil {
		return err
	}

	_, err = r.DatabaseProvider.GetDb().Exec(`DELETE FROM apps WHERE app_id = $1`, appId)
	if err != nil {
		u.Logger.Error("Failed to delete app", deepstack.ErrorField, err)
		return fmt.Errorf("failed to delete app")
	}

	_, err = r.DatabaseProvider.GetDb().Exec("UPDATE users SET used_space_in_bytes = used_space_in_bytes - $1 WHERE user_id = $2", totalDataSize, userId)
	if err != nil {
		return fmt.Errorf("failed to update user space: %w", err)
	}

	return nil
}

// TODO !! or better in user repo?
func (r *AppRepositoryImpl) GetUserIdOfApp(appId int) (int, error) {
	var userId int
	err := r.DatabaseProvider.GetDb().QueryRow(`SELECT user_id FROM apps WHERE app_id = $1`, appId).Scan(&userId)
	if err != nil {
		u.Logger.Error("Failed to get user ID of app", tools.AppIdField, appId, deepstack.ErrorField, err)
		return -1, fmt.Errorf("failed to get user ID of app")
	}
	return userId, nil
}

func (r *AppRepositoryImpl) sumBlobSizes(appID int) (int64, error) {
	var totalSize sql.NullInt64
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT SUM(LENGTH(data)) FROM versions WHERE app_id = $1", appID).Scan(&totalSize)
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

	rows, err := r.DatabaseProvider.GetDb().Query(query, "%"+request.SearchTerm+"%", "%"+request.SearchTerm+"%")
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

// TODO !! dont use DTO
func (r *AppRepositoryImpl) GetAppList(userId int) ([]store.AppDto, error) {
	user, err := r.UserRepo.GetUserById(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	rows, err := r.DatabaseProvider.GetDb().Query("SELECT app_name, app_id FROM apps WHERE user_id = $1", userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get apps: %w", err)
	}
	defer u.Close(rows)

	var apps []store.AppDto
	for rows.Next() {
		var app string
		var appId int
		if err = rows.Scan(&app, &appId); err != nil {
			return nil, fmt.Errorf("failed to scan app: %w", err)
		}
		// TODO !! I think the username is not necessary here, since the use case here is that a user sees his own apps of which he is the maintainer
		apps = append(apps, store.AppDto{Maintainer: user.Name, Name: app, Id: strconv.Itoa(appId)})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return apps, nil
}

func (r *AppRepositoryImpl) GetAppName(appId int) (string, error) {
	var appName string
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT app_name FROM apps WHERE app_id = $1", appId).Scan(&appName)
	if err != nil {
		return "", fmt.Errorf("failed to get app name: %w", err)
	}
	return appName, nil
}

func (r *AppRepositoryImpl) GetMaintainerName(appId int) (string, error) {
	var maintainer string
	err := r.DatabaseProvider.GetDb().QueryRow(`
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

func (r *AppRepositoryImpl) GetAppId(userID int, app string) (int, error) {
	var appID int
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT app_id FROM apps WHERE user_id = $1 AND app_name = $2", userID, app).Scan(&appID)
	if err != nil {
		return 0, fmt.Errorf("app not found: %v", err)
	}
	return appID, nil
}
