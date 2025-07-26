package runner_api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"github.com/nihal-ramaswamy/RunnerIO/internal/services"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// @Summary Find Area for each runner of the group
// @Description Find Area for each runner of the group
// @Tags Runner Group
// @Produce json
// @Success 200 {object} map[string]float64
// @Router /runner/area/{groupCode} [get]
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
		currentMaxTime := time.Unix(0, 0).Nanosecond()
		area, err := services.AreaProcessor(h.ctx, code, h.mongoClient, h.log, currentMaxTime)
		utils.FailIfError(err, c, h.log, http.StatusInternalServerError, "failed to calculate area")

		c.JSON(http.StatusAccepted, area)
	}
}
