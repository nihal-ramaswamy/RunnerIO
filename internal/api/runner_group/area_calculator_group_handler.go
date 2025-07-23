package runner_api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"github.com/nihal-ramaswamy/RunnerIO/internal/services"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type AreaCalculatorGroupHandler struct {
	ctx         context.Context
	mongoClient *mongo.Client
	log         *zap.Logger

	interfaces.HandlerInterface
}

func NewAreaCalculatorGroupHandler(ctx context.Context, mongoClient *mongo.Client, log *zap.Logger) *CreateRunnerGroupHandler {
	return &CreateRunnerGroupHandler{
		mongoClient: mongoClient,
		log:         log,
		ctx:         ctx,
	}
}

func (*AreaCalculatorGroupHandler) Pattern() string {
	return "/area/:groupCode"
}

func (*AreaCalculatorGroupHandler) Method() string {
	return http.MethodGet
}

func (h *AreaCalculatorGroupHandler) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}

func (h *AreaCalculatorGroupHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("groupCode")
		area, err := services.AreaProcessor(h.ctx, code, h.mongoClient, h.log)
		utils.FailIfError(err, c, h.log, http.StatusInternalServerError, "failed to calculate area")

		c.JSON(http.StatusAccepted, area)
	}
}
