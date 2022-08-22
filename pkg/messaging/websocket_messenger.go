package messaging

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
)

type WebsocketMessenger struct {
	connection *websocket.Conn
	sendMu     sync.Mutex
	recvMu     sync.Mutex
}

func NewWebsocketMessenger(conn *websocket.Conn) *WebsocketMessenger {
	handler := new(WebsocketMessenger)
	handler.connection = conn

	return handler
}

func (w *WebsocketMessenger) Send(message Message) error {
	w.sendMu.Lock()
	defer w.sendMu.Unlock()
	msgType, rawBytes, err := Pack(message)
	if err != nil {
		return err
	}

	return w.connection.WriteMessage(msgType, rawBytes)
}

func (w *WebsocketMessenger) Receive() (Message, error) {
	w.recvMu.Lock()
	defer w.recvMu.Unlock()
	msgType, rawBytes, err := w.connection.ReadMessage()
	if err != nil {
		return Message{}, err
	}

	if msgType != websocket.TextMessage {
		return Message{}, errors.New("only text messages are supported")
	}

	message, err := Unpack(rawBytes)
	if err != nil {
		return Message{}, err
	}

	return message, nil

}
