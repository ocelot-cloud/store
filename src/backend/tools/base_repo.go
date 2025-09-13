package tools

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
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
	host = "ocelotcloud_store_postgres"
	customPostgresPort = "5432"

	d.Db, err = u.WaitForPostgresDb(host, customPostgresPort)
	if err != nil {
		return err
	}

	err = u.RunMigrations(d.PathProvider.GetMigrationsDir(), host, customPostgresPort)
	if err != nil {
		return err
	}
	return nil
}
