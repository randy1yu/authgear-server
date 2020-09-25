package service

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AppService), "*"),
	wire.Struct(new(AdminAPIService), "*"),
	wire.Struct(new(AuthzService), "*"),
	wire.Struct(new(ConfigService), "*"),
	NewConfigServiceLogger,
	NewAppServiceLogger,

	wire.Bind(new(AppAuthzService), new(*AuthzService)),
	wire.Bind(new(AppConfigService), new(*ConfigService)),
	wire.Bind(new(AppAdminAPIService), new(*AdminAPIService)),
	wire.Bind(new(AuthzConfigService), new(*ConfigService)),
)
