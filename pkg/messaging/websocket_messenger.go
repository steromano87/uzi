package messaging

import (
	"encoding/json"
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
	rawBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return w.connection.WriteMessage(websocket.TextMessage, rawBytes)
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

	var message Message
	err = json.Unmarshal(rawBytes, &message)
	if err != nil {
		return Message{}, err
	}

	return message, nil
}

func (w *WebsocketMessenger) SendPing() error {
	// TODO: verify how to set ping message deadline to use the more appropriate `WriteControl` function
	return w.Send(NewPingMessage())
}

func (w *WebsocketMessenger) SendPong(pingMsgID string) error {
	return w.Send(NewPongMessage(pingMsgID))
}

func (w *WebsocketMessenger) Close() {
	//TODO implement me
	panic("implement me")
}
