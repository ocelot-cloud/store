//go:build wireinject

package main

import "github.com/google/wire"

func WireDependencies() *InitializerDependencies {
	wire.Build(
		wire.Struct(new(InitializerDependencies), "*"),
	)
	return nil
}

type InitializerDependencies struct{}
