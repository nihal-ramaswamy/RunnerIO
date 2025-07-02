package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	healthcheck_api "github.com/nihal-ramaswamy/RunnerIO/internal/api/healthCheck"
	runner_api "github.com/nihal-ramaswamy/RunnerIO/internal/api/runner"
	amqpconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/amqp"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func NewRoutes(
	server *gin.Engine,
	log *zap.Logger,
	ctx context.Context,
	redisClient *redis.Client,
	mongoClient *mongo.Client,
	ampqconfig *amqpconfig.AmqpConfig,
) {
	serverGroupHandlers := []interfaces.ServerGroupInterface{
		healthcheck_api.NewHealthCheckGroup(ctx, redisClient, log),
		runner_api.NewRunnerGroup(ctx, redisClient, log, mongoClient, ampqconfig),
	}

	for _, serverGroupHandler := range serverGroupHandlers {
		newGroup(server, serverGroupHandler)
	}
}

func newGroup(server *gin.Engine, groupHandler interfaces.ServerGroupInterface) {
	group := server.Group(groupHandler.Group(), groupHandler.Middlewares()...)
	{
		for _, route := range groupHandler.RouteHandlers() {
			newRoute(group, route)
		}
	}
}

func newRoute(server *gin.RouterGroup, routeHandler interfaces.HandlerInterface) {
	middlewares := routeHandler.Middlewares()
	middlewares = append(middlewares, routeHandler.Handler())
	switch routeHandler.RequestMethod() {
	case http.MethodGet:
		server.GET(routeHandler.Pattern(), middlewares...)
	case http.MethodPost:
		server.POST(routeHandler.Pattern(), middlewares...)
	}
}
