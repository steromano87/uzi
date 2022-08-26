package messaging

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"sync"
)

type WebsocketMessenger struct {
	ctx context.Context

	connection *websocket.Conn
	sendMu     sync.Mutex
	recvMu     sync.Mutex
	sendChan   chan Message
	recvChan   chan Message

	closeChan chan struct{}
}

func NewWebsocketMessenger(conn *websocket.Conn, channelBuffer int) *WebsocketMessenger {
	messenger := new(WebsocketMessenger)
	messenger.connection = conn
	messenger.sendChan = make(chan Message, channelBuffer)
	messenger.recvChan = make(chan Message, channelBuffer)

	return messenger
}

func (w *WebsocketMessenger) Start(ctx context.Context) {
	w.ctx = ctx
	go w.pumpIncomingMessages()
	go w.pumpOutgoingMessages()
}

func (w *WebsocketMessenger) Send(message Message) {
	w.sendChan <- message
}

func (w *WebsocketMessenger) Receive() <-chan Message {
	return w.recvChan
}

func (w *WebsocketMessenger) SendPing() {
	w.Send(NewPingMessage())
}

func (w *WebsocketMessenger) SendPong(pingMsgID string) {
	w.Send(NewPongMessage(pingMsgID))
}

func (w *WebsocketMessenger) Close() {
	w.closeChan <- struct{}{}
}

func (w *WebsocketMessenger) pumpIncomingMessages() {
	for {
		_, rawBytes, err := w.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// TODO: add logs implementation
				break
			}
		}

		// TODO: add different messages handling (ping, pong, closure, etc)
		var message Message
		err = json.Unmarshal(rawBytes, &message)
		if err != nil {
			continue
		}

		w.recvChan <- message
	}
}

func (w *WebsocketMessenger) pumpOutgoingMessages() {
	select {
	case <-w.ctx.Done():
		return
	case <-w.closeChan:
		return
	case message := <-w.sendChan:
		rawBytes, err := json.Marshal(message)
		if err != nil {
			break
		}

		var wsMessageType int
		switch message.Type {
		case PingMsgId:
			wsMessageType = websocket.PingMessage
		case PongMsgId:
			wsMessageType = websocket.PongMessage
		default:
			wsMessageType = websocket.TextMessage
		}

		err = w.connection.WriteMessage(wsMessageType, rawBytes)
		if err != nil {
			break
		}
	}
}
