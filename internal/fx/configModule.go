package fx_utils

import (
	redisconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/redis"
	serverconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/server"
	"go.uber.org/fx"
)

var ConfigModule = fx.Module(
	"Config",
	fx.Provide(serverconfig.Default),
	fx.Provide(redisconfig.DefaultRedisConfig),
)
