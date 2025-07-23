package healthcheck_api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"go.uber.org/zap"
)

type HealthCheckHandler struct {
	middlewares []gin.HandlerFunc

	interfaces.HandlerInterface
}

func NewHealthCheckHandler(ctx context.Context, log *zap.Logger) *HealthCheckHandler {
	return &HealthCheckHandler{
		middlewares: []gin.HandlerFunc{},
	}
}

func (*HealthCheckHandler) Pattern() string {
	return "/healthcheck"
}

func (*HealthCheckHandler) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	}
}

func (*HealthCheckHandler) RequestMethod() string {
	return http.MethodGet
}

func (h *HealthCheckHandler) Middlewares() []gin.HandlerFunc {
	return h.middlewares
}
