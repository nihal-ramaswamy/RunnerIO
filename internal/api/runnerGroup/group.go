package runner_api

import (
	"context"

	"github.com/gin-gonic/gin"
	amqpconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/amqp"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	auth_middleware "github.com/nihal-ramaswamy/RunnerIO/internal/middlewares/auth"
)

type RunnerGroup struct {
	routeHandlers []interfaces.HandlerInterface
	middlewares   []gin.HandlerFunc

	interfaces.ServerGroupInterface
}

func (h *RunnerGroup) Group() string {
	return "/runner"
}

func (h *RunnerGroup) RouteHandlers() []interfaces.HandlerInterface {
	return h.routeHandlers
}

func NewRunnerGroup(
	ctx context.Context,
	redisClient *redis.Client,
	log *zap.Logger,
	mongoClient *mongo.Client,
	ampqconfig *amqpconfig.AmqpConfig) *RunnerGroup {
	handlers := []interfaces.HandlerInterface{
		NewCreateRunnerGroupHandler(ctx, mongoClient, log),
		NewJoinRunnerGroupHandler(ctx, mongoClient, log),
	}

	return &RunnerGroup{
		routeHandlers: handlers,
		middlewares: []gin.HandlerFunc{
			auth_middleware.UserInfoMiddleware(ctx, redisClient, log),
		},
	}
}

func (*RunnerGroup) AuthRequired() bool {
	return true
}

func (h *RunnerGroup) Middlewares() []gin.HandlerFunc {
	return h.middlewares
}
