package network

import (
	"github.com/gorilla/websocket"
	"sync"
)

type WebsocketMessenger struct {
	connection *websocket.Conn
	mu         sync.Mutex
}

func NewWebsocketMessenger(conn *websocket.Conn) *WebsocketMessenger {
	handler := new(WebsocketMessenger)
	handler.connection = conn

	return handler
}

func (h *WebsocketMessenger) Write(messageType int, packedMessage []byte) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.connection.WriteMessage(messageType, packedMessage)
}

func (h *WebsocketMessenger) Read() (int, []byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.connection.ReadMessage()
}
