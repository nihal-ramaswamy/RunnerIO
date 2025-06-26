package healthcheck_api

import (
	"context"

	"github.com/gin-gonic/gin"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"go.uber.org/zap"
)

type HealthCheckGroup struct {
	routeHandlers []interfaces.HandlerInterface
	middlewares   []gin.HandlerFunc

	interfaces.ServerGroupInterface
}

func (*HealthCheckGroup) Group() string {
	return "/healthcheck"
}

func (h *HealthCheckGroup) RouteHandlers() []interfaces.HandlerInterface {
	return h.routeHandlers
}

func NewHealthCheckGroup(
	ctx context.Context,
	log *zap.Logger) *HealthCheckGroup {
	handlers := []interfaces.HandlerInterface{
		NewHealthCheckHandler(ctx, log),
		NewAuthHealthCheckHandler(ctx, log),
	}

	return &HealthCheckGroup{
		routeHandlers: handlers,
		middlewares:   []gin.HandlerFunc{},
	}
}

func (*HealthCheckGroup) AuthRequired() bool {
	return false
}

func (h *HealthCheckGroup) Middlewares() []gin.HandlerFunc {
	return h.middlewares
}
