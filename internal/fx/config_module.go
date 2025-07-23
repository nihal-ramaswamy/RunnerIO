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
	fx.Provide(redisconfig.DefaultRedisConfigForAuth),
	fx.Provide(mongodbconfig.DefaultConfig),
	fx.Provide(amqpconfig.DefaultAmqpConfig),
	fx.Provide(wsconfig.DefaultWsConfig),
)
