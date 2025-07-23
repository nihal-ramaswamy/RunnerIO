package fx_utils

import (
	mongodbconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/mongodb"
	redisconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/redis"
	"go.uber.org/fx"
)

var DBModule = fx.Module(
	"DB",
	fx.Provide(redisconfig.NewRedisClient),
	fx.Provide(mongodbconfig.Connect),
)
