package websocket

import (
	"github.com/gorilla/websocket"
)

// Upgrader is the WebSocket upgrader
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Implement proper origin checking in production
	},
}
