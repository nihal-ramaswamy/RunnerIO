package runner_api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	amqpconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/amqp"
	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	"github.com/nihal-ramaswamy/RunnerIO/internal/dto"
	dtoschema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/schema"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
	"go.uber.org/zap"
)

type PutOnQueueGroupHandler struct {
	amqpConfig *amqpconfig.AmqpConfig
	log        *zap.Logger
	interfaces.HandlerInterface
}

func NewPutOnQueueGroupHandler(amqpConfig *amqpconfig.AmqpConfig,
	log *zap.Logger) *PutOnQueueGroupHandler {
	return &PutOnQueueGroupHandler{
		amqpConfig: amqpConfig,
		log:        log,
	}
}

func (*PutOnQueueGroupHandler) Pattern() string {
	return "/putOnQueueGroup"
}

func (h *PutOnQueueGroupHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestData dtoschema.RunnerAuditSchemaRequest

		userDataStr := c.GetHeader("userData")
		var userData dto.UserData
		err := json.Unmarshal([]byte(userDataStr), &userData)
		utils.FailIfError(err, c, h.log, http.StatusInternalServerError,
			"Failed to unmarshal user data", zap.Error(err))

		err = c.ShouldBindJSON(&requestData)
		utils.FailIfError(err, c, h.log, http.StatusBadRequest,
			"Failed to bind JSON", zap.Error(err))

		data := requestData.ToRunnerAuditSchema(userData.Sub)

		// Validate GroupCode - Ensure it's not empty
		if requestData.GroupCode == "" {
			h.log.Error("Group code is missing in request", zap.Any("requestData", requestData))
			c.AbortWithError(http.StatusBadRequest, errors.New("group code is required"))
			return
		}

		// Prepare the message for RabbitMQ
		jsonData, err := json.Marshal(data)
		utils.FailIfError(err, c, h.log, http.StatusInternalServerError,
			"Failed to marshal request data", zap.Error(err))

		err = h.amqpConfig.PublishWithContext(jsonData, constants.AUDIT_QUEUE_NAME)
		utils.FailIfError(err, c, h.log, http.StatusInternalServerError,
			"Failed to publish message to RabbitMQ",
			zap.Error(err), zap.String("queue", constants.AUDIT_QUEUE_NAME))

		c.JSON(http.StatusOK, gin.H{
			"message": "Points added to queue successfully",
		})
	}
}

func (*PutOnQueueGroupHandler) RequestMethod() string {
	return http.MethodPost
}

func (h *PutOnQueueGroupHandler) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}
