//go:build wireinject

package setup

import (
	"net/http"
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
		tools.NewConfig,
		NewPathProvider,
		NewMux,

		wire.Struct(new(InitializerDependencies), "*"),
		wire.Struct(new(apps.AppsHandler), "*"),
		wire.Struct(new(apps.AppRepositoryImpl), "*"),
		wire.Struct(new(HandlerInitializer), "*"),
		wire.Struct(new(versions.VersionsHandler), "*"),
		wire.Struct(new(versions.VersionRepositoryImpl), "*"),
		wire.Struct(new(users.UserRepositoryImpl), "*"),
		wire.Struct(new(users.UserHandler), "*"),
		wire.Struct(new(users.EmailClientImpl), "*"),
		wire.Struct(new(users.EmailConfigStoreImpl), "*"),
		wire.Struct(new(versions.VersionService), "*"),
		wire.Struct(new(Server), "*"),
		wire.Struct(new(users.UserServiceImpl), "*"),
		wire.Struct(new(apps.AppServiceImpl), "*"),
		wire.Struct(new(users.RegistrationCodeProvider), "*"),

		wire.Bind(new(apps.AppRepository), new(*apps.AppRepositoryImpl)),
		wire.Bind(new(versions.VersionRepository), new(*versions.VersionRepositoryImpl)),
		wire.Bind(new(users.UserRepository), new(*users.UserRepositoryImpl)),
	)
	return nil
}

type InitializerDependencies struct {
	HandlerInitializer *HandlerInitializer
	DatabaseProvider   *tools.DatabaseProviderImpl
	PathProvider       *tools.PathProviderImpl
	Server             *Server
}

func NewDatabaseProvider(pathProvider *tools.PathProviderImpl) *tools.DatabaseProviderImpl {
	return &tools.DatabaseProviderImpl{
		Db:           nil,
		PathProvider: pathProvider,
	}
}

func NewEmailVerifier() *tools.EmailVerifierImpl {
	return &tools.EmailVerifierImpl{
		WaitingList: sync.Map{},
	}
}

func NewPathProvider() *tools.PathProviderImpl {
	return &tools.PathProviderImpl{}
}

func NewMux() *http.ServeMux {
	return http.NewServeMux()
}
