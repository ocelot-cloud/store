package tools

import (
	"database/sql"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ocelot-cloud/deepstack"
	u "github.com/ocelot-cloud/shared/utils"
)

type DatabaseProviderImpl struct {
	Db           *sql.DB
	PathProvider *PathProviderImpl
}

func (d *DatabaseProviderImpl) GetDb() *sql.DB {
	return d.Db
}

func (d *DatabaseProviderImpl) InitializeDatabase() error {
	var err error
	var host, customPostgresPort string
	// TODO maybe better introduce profiles? -> so acceptance testing should be run against the app store container I guess?
	host = "ocelotcloud_store_postgres"
	customPostgresPort = "5432"

	// TODO !! get rid of exit() and return error to top module instead

	d.Db, err = u.WaitForPostgresDb(host, customPostgresPort)
	if err != nil {
		u.Logger.Error("Failed to create database client", deepstack.ErrorField, err)
		os.Exit(1)
	}
	u.Logger.Info("Database client created successfully")

	err = u.RunMigrations(d.PathProvider.GetMigrationsDir(), host, customPostgresPort)
	if err != nil {
		u.Logger.Error("Failed to run migrations", deepstack.ErrorField, err)
		os.Exit(1)
	}
	return nil
}
