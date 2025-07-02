package wsconfig

import (
	"github.com/gorilla/websocket"
)

func DefaultWsConfig() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
}
