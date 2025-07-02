package runner_ws

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	amqpconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/amqp"
	"github.com/nihal-ramaswamy/RunnerIO/internal/constants"
	"github.com/nihal-ramaswamy/RunnerIO/internal/dto"
	dtoschema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/schema"
	wsdto "github.com/nihal-ramaswamy/RunnerIO/internal/dto/ws"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
	"go.uber.org/zap"
)

type PersistAuditDataGroupHandler struct {
	amqpConfig                 *amqpconfig.AmqpConfig
	log                        *zap.Logger
	upgrader                   *websocket.Upgrader
	persistAuditDataClientsMap *wsdto.PersistAuditDataManagerMap

	interfaces.HandlerInterface
}

func NewPutOnQueueGroupHandler(
	amqpConfig *amqpconfig.AmqpConfig,
	log *zap.Logger,
	upgrader *websocket.Upgrader,
	persistAuditDataClientsMap *wsdto.PersistAuditDataManagerMap,
) *PersistAuditDataGroupHandler {
	return &PersistAuditDataGroupHandler{
		amqpConfig:                 amqpConfig,
		log:                        log,
		upgrader:                   upgrader,
		persistAuditDataClientsMap: persistAuditDataClientsMap,
	}
}

func (*PersistAuditDataGroupHandler) Pattern() string {
	return "/ws/putOnQueueGroup/"
}

func (h *PersistAuditDataGroupHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ws, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
		utils.FailIfError(err, c, h.log, http.StatusInternalServerError,
			"Failed to upgrade connection", zap.Error(err))
		defer ws.Close()

		var requestData dtoschema.RunnerAuditSchemaRequest

		userDataStr := c.GetHeader("userData")
		var userData dto.UserData
		err = json.Unmarshal([]byte(userDataStr), &userData)
		utils.FailIfError(err, c, h.log, http.StatusInternalServerError,
			"Failed to unmarshal user data", zap.Error(err))

		h.persistAuditDataClientsMap.Add(userData.Sub, &wsdto.PersistAuditDataClient{
			Conn: ws,
			Sub:  userData.Sub,
		})

		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					h.log.Error("Unexpected close error", zap.Error(err))
					break
				}
				h.log.Error("Failed to read message", zap.Error(err))
			}

			err = json.Unmarshal(message, &requestData)
			utils.FailIfError(err, c, h.log, http.StatusBadRequest,
				"Incorrect message format", zap.Error(err))

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

			client, ok := h.persistAuditDataClientsMap.Get(userData.Sub)
			if !ok {
				h.log.Error("Client not found", zap.String("sub", userData.Sub))
				continue
			}

			resp, err := json.Marshal(gin.H{
				"message": "Points added to queue successfully",
			})

			utils.FailIfError(err, c, h.log, http.StatusInternalServerError,
				"Failed to marshal response data", zap.Error(err))

			client.Conn.WriteMessage(websocket.TextMessage, resp)

			c.JSON(http.StatusOK, gin.H{
				"message": "Points added to queue successfully",
			})
		}
	}
}

func (*PersistAuditDataGroupHandler) RequestMethod() string {
	return http.MethodGet
}

func (h *PersistAuditDataGroupHandler) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}
