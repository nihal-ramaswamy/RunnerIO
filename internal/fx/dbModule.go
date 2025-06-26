package fx_utils

import (
	redisconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/redis"
	"go.uber.org/fx"
)

var DBModule = fx.Module(
	"DB",
	fx.Provide(redisconfig.NewRedisClient),
)
