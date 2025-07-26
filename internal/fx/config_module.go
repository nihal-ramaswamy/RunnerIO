package fx_utils

import (
	amqpconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/amqp"
	mongodbconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/mongodb"
	redisconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/redis"
	serverconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/server"
	wsconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/ws"
	"go.uber.org/fx"
)

var ConfigModule = fx.Module(
	"Config",
	fx.Provide(serverconfig.Default),
	fx.Provide(
		fx.Annotate(
			redisconfig.DefaultRedisConfigForAuth,
			fx.ResultTags(`name:"auth_rdb_config"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			redisconfig.DefaultRedisConfigForEngine,
			fx.ResultTags(`name:"engine_rdb_config"`),
		),
	),
	fx.Provide(mongodbconfig.DefaultConfig),
	fx.Provide(amqpconfig.DefaultAmqpConfig),
	fx.Provide(wsconfig.DefaultWsConfig),
)
