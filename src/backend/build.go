//go:build wireinject

package main

import (
	"ocelot/store/apps"
	"ocelot/store/versions"

	"github.com/google/wire"
)

func WireDependencies() *InitializerDependencies {
	wire.Build(
		wire.Struct(new(InitializerDependencies), "*"),
		wire.Struct(new(apps.AppsHandler), "*"),
		wire.Struct(new(apps.AppRepositoryImpl), "*"),
		wire.Struct(new(HandlerInitializer), "*"),
		wire.Struct(new(versions.VersionsHandler), "*"),
		wire.Struct(new(DatabaseSampleDataSeeder), "*"),
		wire.Struct(new(versions.VersionRepositoryImpl), "*"),

		wire.Bind(new(apps.AppRepository), new(*apps.AppRepositoryImpl)),
		wire.Bind(new(versions.VersionRepository), new(*versions.VersionRepositoryImpl)),
	)
	return nil
}

type InitializerDependencies struct {
	HandlerInitializer *HandlerInitializer
}
