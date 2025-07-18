package tools

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ocelot-cloud/shared/utils"
	"os"
	"sync"
)

var Db *sql.DB

func InitializeDatabase() {
	var err error
	var host, customPostgresPort string
	if os.Getenv("RUN_NATIVELY") == "true" {
		host = "localhost"
		customPostgresPort = "5433"
	} else {
		host = "ocelotcloud_store_postgres"
		customPostgresPort = "5432"
	}

	Db, err = utils.WaitForPostgresDb(host, customPostgresPort)
	if err != nil {
		Logger.ErrorF("Failed to create database client: %v", err)
		os.Exit(1)
	}

	migrationsDir := utils.FindDir("assets") + "/migrations"
	utils.RunMigrations(migrationsDir, host, customPostgresPort)
}

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
