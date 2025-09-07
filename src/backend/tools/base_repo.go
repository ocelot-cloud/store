package tools

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ocelot-cloud/deepstack"
	"github.com/ocelot-cloud/shared/utils"
)

// TODO !! global var
var Db *sql.DB

// TODO !! should return error
func InitializeDatabase() {
	var err error
	var host, customPostgresPort string
	// TODO maybe better introduce profiles? -> so acceptance testing should be run against the app store container I guess?
	if Profile == TEST {
		host = "localhost"
		customPostgresPort = "5433"
	} else {
		// TODO !! always use this one
		host = "ocelotcloud_store_postgres"
		customPostgresPort = "5432"
	}

	// TODO !! get rid of exit() and return error to top module instead

	Db, err = utils.WaitForPostgresDb(host, customPostgresPort)
	if err != nil {
		Logger.Error("Failed to create database client", deepstack.ErrorField, err)
		os.Exit(1)
	}

	// TODO !! directories and path initialization are two separate concerns
	assertDir, err := utils.FindDir("assets")
	if err != nil {
		Logger.Error("Failed to find migrations directory", deepstack.ErrorField, err)
		os.Exit(1)
	}
	migrationsDir := assertDir + "/migrations"

	err = utils.RunMigrations(migrationsDir, host, customPostgresPort)
	if err != nil {
		Logger.Error("Failed to run migrations", deepstack.ErrorField, err)
		os.Exit(1)
	}

}

// TODO !! global var
var WaitingForEmailVerificationList sync.Map

func GetAppId(userID int, app string) (int, error) {
	var appID int
	err := Db.QueryRow("SELECT app_id FROM apps WHERE user_id = $1 AND app_name = $2", userID, app).Scan(&appID)
	if err != nil {
		return 0, fmt.Errorf("app not found: %v", err)
	}
	return appID, nil
}

func GetUserId(user string) (int, error) {
	var userID int
	err := Db.QueryRow("SELECT user_id FROM users WHERE user_name = $1", user).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("user not found: %w", err)
	}
	return userID, nil
}
