//go:build wireinject

package main

import (
	"ocelot/store/apps"
	"ocelot/store/tools"
	"ocelot/store/users"
	"ocelot/store/versions"
	"sync"

	"github.com/google/wire"
)

func WireDependencies() *InitializerDependencies {
	wire.Build(
		NewDatabaseProvider,
		NewEmailVerifier,
		
		wire.Struct(new(InitializerDependencies), "*"),
		wire.Struct(new(apps.AppsHandler), "*"),
		wire.Struct(new(apps.AppRepositoryImpl), "*"),
		wire.Struct(new(HandlerInitializer), "*"),
		wire.Struct(new(versions.VersionsHandler), "*"),
		wire.Struct(new(DatabaseSampleDataSeeder), "*"),
		wire.Struct(new(versions.VersionRepositoryImpl), "*"),
		wire.Struct(new(users.UserRepositoryImpl), "*"),
		wire.Struct(new(users.UserHandler), "*"),

		wire.Bind(new(apps.AppRepository), new(*apps.AppRepositoryImpl)),
		wire.Bind(new(versions.VersionRepository), new(*versions.VersionRepositoryImpl)),
		wire.Bind(new(users.UserRepository), new(*users.UserRepositoryImpl)),
	)
	return nil
}

type InitializerDependencies struct {
	HandlerInitializer *HandlerInitializer
	DatabaseProvider   *tools.DatabaseProviderImpl
}

func NewDatabaseProvider() *tools.DatabaseProviderImpl {
	return &tools.DatabaseProviderImpl{}
}

func NewEmailVerifier() *tools.EmailVerifierImpl {
	return &tools.EmailVerifierImpl{
		WaitingList: sync.Map{},
	}
}
