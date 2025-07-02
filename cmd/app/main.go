package main

import (
	"context"

	"github.com/gin-gonic/gin"
	amqpconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/amqp"
	serverconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/server"
	fx_utils "github.com/nihal-ramaswamy/RunnerIO/internal/fx"
	"github.com/nihal-ramaswamy/RunnerIO/internal/messages"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(utils.NewProduction),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),

		fx_utils.ConfigModule,
		fx_utils.DBModule,
		fx_utils.MicroServicesModule,

		fx.Invoke(Invoke),
	).Run()
}

func Invoke(server *gin.Engine, config *serverconfig.Config, log *zap.Logger, ctx context.Context, amqpconfig *amqpconfig.AmqpConfig, mongoClient *mongo.Client) {
	go func() {
		err := server.Run(config.Port)
		if nil != err {
			log.Error(err.Error())
		}
	}()
	go func() {
		log.Info("Starting Audit Consumer")
		messages.PersistAuditData(ctx, amqpconfig, mongoClient, log)
	}()

}
