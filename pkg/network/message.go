package network

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"time"
)

type Message struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Payload   any       `json:"payload"`
}

func NewMessage(msgType string, payload any) *Message {
	message := new(Message)
	ID, _ := uuid.NewRandom()
	message.ID = ID.String()
	message.Timestamp = time.Now()
	message.Type = msgType
	message.Payload = payload

	return message
}

func (m Message) WebSocketType() int {
	if _, ok := m.Payload.([]byte); ok {
		return websocket.BinaryMessage
	}

	return websocket.TextMessage
}

func (m Message) PayloadMap() (map[string]any, error) {
	if m.WebSocketType() == websocket.BinaryMessage {
		return nil, errors.New("binary payload cannot be converted into a map")
	}

	mappedPayload, ok := m.Payload.(map[string]any)
	if !ok {
		return nil, errors.New(fmt.Sprintf("error during payload conversion to map: %+v", m.Payload))
	}

	return mappedPayload, nil
}
