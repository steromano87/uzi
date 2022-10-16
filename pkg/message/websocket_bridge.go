package message

import (
	"context"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"sync"
)

type WebsocketBridge struct {
	ctx context.Context

	connection *websocket.Conn
	sendMu     sync.Mutex
	recvMu     sync.Mutex
	sendChan   chan *Envelope
	recvChan   chan *Envelope

	closeChan chan struct{}
}

func NewWebsocketBridge(conn *websocket.Conn, channelBuffer int) *WebsocketBridge {
	bridge := new(WebsocketBridge)
	bridge.connection = conn
	bridge.sendChan = make(chan *Envelope, channelBuffer)
	bridge.recvChan = make(chan *Envelope, channelBuffer)

	return bridge
}

func (w *WebsocketBridge) Start(ctx context.Context) {
	w.ctx = ctx
	go w.pumpIncomingMessages()
	go w.pumpOutgoingMessages()
}

func (w *WebsocketBridge) Send(message *Envelope) {
	w.sendChan <- message
}

func (w *WebsocketBridge) Receive() <-chan *Envelope {
	return w.recvChan
}

func (w *WebsocketBridge) SendPing() {
	payload := Envelope_Ping{Ping: &Ping{}}
	msg := NewEnvelope(&payload)
	w.Send(msg)
}

func (w *WebsocketBridge) SendPong(pingMsgID string) {
	payload := Envelope_Pong{Pong: &Pong{}}
	msg := NewResponseEnvelope(pingMsgID, &payload)
	w.Send(msg)
}

func (w *WebsocketBridge) Close() {
	w.closeChan <- struct{}{}
}

func (w *WebsocketBridge) pumpIncomingMessages() {
	for {
		_, rawBytes, err := w.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// TODO: add logs implementation
				break
			}
		}

		// TODO: add different messages handling (ping, pong, closure, etc)
		var envelope Envelope
		err = proto.Unmarshal(rawBytes, &envelope)
		if err != nil {
			continue
		}

		w.recvChan <- &envelope
	}
}

func (w *WebsocketBridge) pumpOutgoingMessages() {
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
		case *Envelope_Ping:
			wsMessageType = websocket.PingMessage
		case *Envelope_Pong:
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
