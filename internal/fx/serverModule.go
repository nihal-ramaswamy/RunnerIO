package fx_utils

import (
	"context"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nihal-ramaswamy/RunnerIO/internal/api"
	serverconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/server"
	log_middleware "github.com/nihal-ramaswamy/RunnerIO/internal/middlewares/log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func newServerEngine(
	lc fx.Lifecycle,
	config *serverconfig.Config,
	log *zap.Logger,
	ctx context.Context,
	redisClient *redis.Client,
) *gin.Engine {
	gin.SetMode(config.GinMode)

	server := gin.Default()

	cfg := cors.DefaultConfig()
	cfg.AllowAllOrigins = true
	cfg.AllowMethods = []string{"POST", "GET", "PUT", "OPTIONS"}
	cfg.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "User-Agent", "Cache-Control", "Pragma"}
	cfg.ExposeHeaders = []string{"Content-Length"}
	cfg.AllowCredentials = true
	cfg.MaxAge = 12 * time.Hour

	server.Use(cors.New(cfg))
	server.Use(log_middleware.DefaultStructuredLogger(log))
	server.Use(gin.Recovery())

	api.NewRoutes(server, log, ctx, redisClient)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting server on port", zap.String("port", config.Port))

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping server")
			defer func() {
				err := log.Sync()
				if nil != err {
					log.Error(err.Error())
				}
			}()

			return nil
		},
	})

	return server
}

var serverModule = fx.Module(
	"serverModule",
	fx.Provide(
		newServerEngine,
	),
)
