package apps

import (
	"database/sql"
	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/store"
	u "github.com/ocelot-cloud/shared/utils"
	"ocelot/store/tools"
	"ocelot/store/users"
	"strconv"
)

type AppRepository interface {
	DoesAppIdExist(appId int) (bool, error)
	CreateApp(userId int, app string) error
	DeleteApp(appId int) error
	GetAppList(userId int) ([]AppItem, error)
	SearchForApps(searchRequest store.AppSearchRequest) ([]store.AppWithLatestVersion, error) // TODO !! not sure whether it makes sense to maybe improve my search function, like explicitly say have a field for maintainer and app you can search for; if empty, its ignored
	GetAppById(appId int) (*tools.App, error)                                                 // TODO !! dont use DTO, ID should be integer
	DoesAppExist(userID int, app string) (bool, error)
	GetUserIdOfApp(appId int) (int, error)
}

type AppRepositoryImpl struct {
	DatabaseProvider *tools.DatabaseProviderImpl
	UserRepo         users.UserRepository
}

func (r *AppRepositoryImpl) GetAppById(appId int) (*tools.App, error) {
	var app tools.App
	err := r.DatabaseProvider.GetDb().QueryRow(
		`SELECT app_id, user_id, app_name
		 FROM apps
		WHERE app_id = $1`,
		appId,
	).Scan(&app.Id, &app.OwnerId, &app.Name)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return &app, nil
}

func (r *AppRepositoryImpl) CreateApp(userId int, app string) error {
	_, err := r.DatabaseProvider.GetDb().Exec(`INSERT INTO apps (user_id, app_name) VALUES ($1, $2)`, userId, app)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}
	return nil
}

func (r *AppRepositoryImpl) DoesAppIdExist(appId int) (bool, error) {
	var exists bool
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT EXISTS(SELECT 1 FROM apps WHERE app_id = $1)", appId).Scan(&exists)
	if err != nil {
		return false, u.Logger.NewError(err.Error())
	}
	return exists, nil
}

// TODO !! this is business logic
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
		return u.Logger.NewError(err.Error())
	}

	_, err = r.DatabaseProvider.GetDb().Exec("UPDATE users SET used_space_in_bytes = used_space_in_bytes - $1 WHERE user_id = $2", totalDataSize, userId)
	if err != nil {
		return u.Logger.NewError(err.Error())
	}

	return nil
}

func (r *AppRepositoryImpl) GetUserIdOfApp(appId int) (int, error) {
	var userId int
	err := r.DatabaseProvider.GetDb().QueryRow(`SELECT user_id FROM apps WHERE app_id = $1`, appId).Scan(&userId)
	if err != nil {
		return -1, u.Logger.NewError(err.Error())
	}
	return userId, nil
}

func (r *AppRepositoryImpl) sumBlobSizes(appID int) (int64, error) {
	var totalSize sql.NullInt64
	err := r.DatabaseProvider.GetDb().QueryRow("SELECT SUM(LENGTH(data)) FROM versions WHERE app_id = $1", appID).Scan(&totalSize)
	if err != nil {
		return 0, u.Logger.NewError(err.Error())
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
		return nil, u.Logger.NewError(err.Error())
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
		return nil, u.Logger.NewError(err.Error())
	}
	return apps, nil
}

type AppItem struct {
	Id   int
	Name string
}

func (r *AppRepositoryImpl) GetAppList(userId int) ([]AppItem, error) {
	rows, err := r.DatabaseProvider.GetDb().Query("SELECT app_name, app_id FROM apps WHERE user_id = $1", userId)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(rows)

	var apps []AppItem
	for rows.Next() {
		// TODO !! simplify
		var name string
		var id int
		if err = rows.Scan(&name, &id); err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		apps = append(apps, AppItem{Name: name, Id: id})
	}

	if err = rows.Err(); err != nil {
		return nil, u.Logger.NewError(err.Error())
	}

	return apps, nil
}

func (r *AppRepositoryImpl) DoesAppExist(userID int, appName string) (bool, error) {
	var exists bool
	err := r.DatabaseProvider.GetDb().
		QueryRow("SELECT EXISTS(SELECT 1 FROM apps WHERE user_id = $1 AND app_name = $2)", userID, appName).
		Scan(&exists)
	if err != nil {
		return false, u.Logger.NewError(err.Error())
	}
	return exists, nil
}
