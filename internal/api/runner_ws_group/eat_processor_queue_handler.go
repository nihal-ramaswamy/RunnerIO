package runner_ws

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	amqpconfig "github.com/nihal-ramaswamy/RunnerIO/internal/config/amqp"
	"github.com/nihal-ramaswamy/RunnerIO/internal/dto"
	wsdto "github.com/nihal-ramaswamy/RunnerIO/internal/dto/ws"
	interfaces "github.com/nihal-ramaswamy/RunnerIO/internal/interface"
	"github.com/nihal-ramaswamy/RunnerIO/internal/utils"
	"go.uber.org/zap"
)

type EatProcessorQueueHandler struct {
	amqpConfig                    *amqpconfig.AmqpConfig
	log                           *zap.Logger
	upgrader                      *websocket.Upgrader
	groupCodeDataClientManagerMap *wsdto.GroupCodeDataClientManagerMap

	interfaces.HandlerInterface
}

func NewEatProcessorQueueHandler(
	amqpConfig *amqpconfig.AmqpConfig,
	log *zap.Logger,
	upgrader *websocket.Upgrader,
	groupCodeDataClientManagerMap *wsdto.GroupCodeDataClientManagerMap,
) *EatProcessorQueueHandler {
	return &EatProcessorQueueHandler{
		amqpConfig:                    amqpConfig,
		log:                           log,
		upgrader:                      upgrader,
		groupCodeDataClientManagerMap: groupCodeDataClientManagerMap,
	}
}

func (*EatProcessorQueueHandler) Pattern() string {
	return "/eatProcessorQueue/:groupCode"
}

func (h *EatProcessorQueueHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		groupCode := c.Param("groupCode")
		ws, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
		utils.FailIfError(err, c, h.log, http.StatusInternalServerError,
			"Failed to upgrade connection", zap.Error(err))
		defer ws.Close()

		userDataStr := c.GetHeader("userData")
		var userData dto.UserData
		err = json.Unmarshal([]byte(userDataStr), &userData)
		utils.FailIfError(err, c, h.log, http.StatusInternalServerError,
			"Failed to unmarshal user data", zap.Error(err))

		client := &wsdto.GroupCodeDataClient{
			Conn: ws,
			Sub:  userData.Sub,
		}
		h.groupCodeDataClientManagerMap.Add(groupCode, client)
		defer func() {
			h.groupCodeDataClientManagerMap.RemoveClient(groupCode, client)
			ws.Close()
		}()

		h.log.Info("declaring queue", zap.String("key", groupCode))
		err = h.amqpConfig.DeclareAndBindQueue(groupCode, groupCode)
		utils.FailIfError(err, c, h.log, http.StatusInternalServerError,
			"Failed to declare and bind queue", zap.Error(err))

		go func() {
			for {
				msgs, err := h.amqpConfig.Channel.Consume(
					groupCode,    // queue
					userData.Sub, // consumer
					false,        // auto-ack
					false,        // exclusive
					false,        // no-local
					false,        // no-wait
					nil,          // args
				)
				if err != nil {
					h.log.Error("Failed to register a consumer", zap.Error(err))
					return
				}

				for d := range msgs {
					h.log.Info("Received a message", zap.String("message", string(d.Body)))
					clients, ok := h.groupCodeDataClientManagerMap.Get(groupCode)
					if !ok {
						h.log.Error("Failed to get clients for group code", zap.String("groupCode", groupCode))
						return
					}
					for _, client := range clients {
						err := client.Conn.WriteMessage(websocket.TextMessage, d.Body)
						if err != nil {
							h.log.Error("Failed to write message", zap.Error(err))
							h.groupCodeDataClientManagerMap.RemoveClient(groupCode, client)
							client.Conn.Close()
						}
					}
				}
			}
		}()

		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				h.log.Info("Client disconnected", zap.String("groupCode", groupCode), zap.String("sub", userData.Sub))
				break
			}
		}
	}
}

func (*EatProcessorQueueHandler) RequestMethod() string {
	return http.MethodGet
}

func (h *EatProcessorQueueHandler) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{}
}
