package messaging

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/steromano87/harkonnen/v1/pkg/protobuf/message"
	"google.golang.org/protobuf/proto"
	"sync"
)

type WebsocketMessenger struct {
	ctx context.Context

	connection *websocket.Conn
	sendMu     sync.Mutex
	recvMu     sync.Mutex
	sendChan   chan *message.Envelope
	recvChan   chan *message.Envelope

	closeChan chan struct{}
}

func NewWebsocketMessenger(conn *websocket.Conn, channelBuffer int) *WebsocketMessenger {
	messenger := new(WebsocketMessenger)
	messenger.connection = conn
	messenger.sendChan = make(chan *message.Envelope, channelBuffer)
	messenger.recvChan = make(chan *message.Envelope, channelBuffer)

	return messenger
}

func (w *WebsocketMessenger) Start(ctx context.Context) {
	w.ctx = ctx
	go w.pumpIncomingMessages()
	go w.pumpOutgoingMessages()
}

func (w *WebsocketMessenger) Send(message *message.Envelope) {
	w.sendChan <- message
}

func (w *WebsocketMessenger) Receive() <-chan *message.Envelope {
	return w.recvChan
}

func (w *WebsocketMessenger) SendPing() {
	payload := message.Envelope_Ping{Ping: &message.Ping{}}
	msg := message.NewEnvelope(&payload)
	w.Send(msg)
}

func (w *WebsocketMessenger) SendPong(pingMsgID string) {
	payload := message.Envelope_Pong{Pong: &message.Pong{}}
	msg := message.NewResponseEnvelope(pingMsgID, &payload)
	w.Send(msg)
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
		var envelope message.Envelope
		err = proto.Unmarshal(rawBytes, &envelope)
		if err != nil {
			continue
		}

		w.recvChan <- &envelope
	}
}

func (w *WebsocketMessenger) pumpOutgoingMessages() {
	select {
	case <-w.ctx.Done():
		return
	case <-w.closeChan:
		return
	case msg := <-w.sendChan:
		rawBytes, err := proto.Marshal(msg)
		if err != nil {
			break
		}

		var wsMessageType int
		switch msg.Payload.(type) {
		case *message.Envelope_Ping:
			wsMessageType = websocket.PingMessage
		case *message.Envelope_Pong:
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
