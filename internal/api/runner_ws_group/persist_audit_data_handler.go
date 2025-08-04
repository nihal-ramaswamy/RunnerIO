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
	mongo_schema "github.com/nihal-ramaswamy/RunnerIO/internal/dto/mongodb_schema"
	wsdto "github.com/nihal-ramaswamy/RunnerIO/internal/dto/ws"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
	"go.uber.org/zap"
)

// @Summary Persist Audit Data
// @Description Join the websocket to continuously send points to the live lines queue.
// @Tags Runner WS Group
// @Message {object} mongo_schema.RunnerLiveLinesSchemaRequest
// @Router /ws/putOnQueueGroup/ [ws]
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
	return "/putOnQueueGroup/"
}

func (h *PersistAuditDataGroupHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ws, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
		utils.FailIfError(err, c, h.log, http.StatusInternalServerError,
			"Failed to upgrade connection", zap.Error(err))
		defer ws.Close()

		var requestData mongo_schema.RunnerLiveLinesSchemaRequest

		userDataStr := c.GetHeader("userData")
		var userData dto.UserData
		err = json.Unmarshal([]byte(userDataStr), &userData)
		utils.FailIfError(err, c, h.log, http.StatusInternalServerError,
			"Failed to unmarshal user data", zap.Error(err))

		client := &wsdto.PersistAuditDataClient{
			Conn: ws,
			Sub:  userData.Sub,
		}
		h.persistAuditDataClientsMap.Add(userData.Sub, client)
		defer func() {
			h.persistAuditDataClientsMap.RemoveClient(client)
			ws.Close()
		}()

		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				h.log.Info("Client disconnected", zap.String("sub", userData.Sub))
				break
			}
			h.log.Info("Read message", zap.String("message", string(message)))

			err = json.Unmarshal(message, &requestData)
			utils.FailIfError(err, c, h.log, http.StatusBadRequest,
				"Incorrect message format", zap.Error(err))

			data := requestData.ToRunnerLiveLinesSchema(userData.Sub)

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

			// Publish the message to audit queue
			err = h.amqpConfig.PublishWithContext(jsonData, constants.LIVE_LINES_QUEUE_NAME, true)
			utils.FailIfError(err, c, h.log, http.StatusInternalServerError,
				"Failed to publish message to RabbitMQ",
				zap.Error(err), zap.String("queue", constants.LIVE_LINES_QUEUE_NAME))

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

			err = client.Conn.WriteMessage(websocket.TextMessage, resp)
			if err != nil {
				h.log.Error("Failed to write message", zap.Error(err))
				h.persistAuditDataClientsMap.RemoveClient(client)
				client.Conn.Close()
			}

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
