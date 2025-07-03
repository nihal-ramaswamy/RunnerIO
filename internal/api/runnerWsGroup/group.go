package runner_ws

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	amqpconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/amqp"
	wsdto "github.com/nihal-ramaswamy/RunnerIO/internal/dto/ws"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	auth_middleware "github.com/nihal-ramaswamy/RunnerIO/internal/middlewares/auth"
)

type RunnerWsGroup struct {
	routeHandlers []interfaces.HandlerInterface
	middlewares   []gin.HandlerFunc

	interfaces.ServerGroupInterface
}

func (h *RunnerWsGroup) Group() string {
	return "/ws"
}

func (h *RunnerWsGroup) RouteHandlers() []interfaces.HandlerInterface {
	return h.routeHandlers
}

func NewRunnerWsGroup(
	ctx context.Context,
	redisClient *redis.Client,
	upgrader *websocket.Upgrader,
	persistAuditDataClientsMap *wsdto.PersistAuditDataManagerMap,
	groupCodeDataClientManagerMap *wsdto.GroupCodeDataClientManagerMap,
	log *zap.Logger,
	ampqconfig *amqpconfig.AmqpConfig) *RunnerWsGroup {
	handlers := []interfaces.HandlerInterface{
		NewPutOnQueueGroupHandler(ampqconfig, log, upgrader, persistAuditDataClientsMap),
		NewEatProcessorQueueHandler(ampqconfig, log, upgrader, groupCodeDataClientManagerMap),
	}

	return &RunnerWsGroup{
		routeHandlers: handlers,
		middlewares: []gin.HandlerFunc{
			auth_middleware.UserInfoMiddleware(ctx, redisClient, log),
		},
	}
}

func (*RunnerWsGroup) AuthRequired() bool {
	return true
}

func (h *RunnerWsGroup) Middlewares() []gin.HandlerFunc {
	return h.middlewares
}
