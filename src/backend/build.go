package main

import "github.com/google/wire"

func WireDependencies() *InitializerDependencies {
	wire.Build()
	return nil
}

type InitializerDependencies struct {
}
