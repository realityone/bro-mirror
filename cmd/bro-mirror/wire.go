//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/realityone/bro-mirror/internal/conf"
	"github.com/realityone/bro-mirror/internal/data"
	"github.com/realityone/bro-mirror/internal/server"
	"github.com/realityone/bro-mirror/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Mirror, *conf.Data_TencentOSS, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(data.ProviderSet, server.ProviderSet, service.ProviderSet, newApp))
}
