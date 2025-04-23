package tools

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ocelot-cloud/shared/utils"
	"sync"
)

var Db *sql.DB

func InitializeDatabase() {
	var err error
	customPostgresPort := "5433"
	Db, err = utils.WaitForPostgresDb("localhost", customPostgresPort)
	if err != nil {
		Logger.Fatal("Failed to create database client: %v", err)
	}

	migrationsDir := utils.FindDir("assets") + "/migrations"
	utils.RunMigrations(migrationsDir, "localhost", customPostgresPort)
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
