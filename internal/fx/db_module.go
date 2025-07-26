package fx_utils

import (
	mongodbconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/mongodb"
	redisconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/redis"
	"go.uber.org/fx"
)

var DBModuleService = fx.Module(
	"DB_Service",
	fx.Provide(
		fx.Annotate(
			redisconfig.NewRedisClient,
			fx.ParamTags(``, ``, `name:"auth_rdb_config"`),
			fx.ResultTags(`name:"auth_rdb"`),
		),
	),
	fx.Provide(mongodbconfig.Connect),
)

var DBModuleEngine = fx.Module(
	"DB_Engine",
	fx.Provide(
		fx.Annotate(
			redisconfig.NewRedisClient,
			fx.ParamTags(``, ``, `name:"engine_rdb_config"`),
		),
	),
	fx.Provide(mongodbconfig.Connect),
)
