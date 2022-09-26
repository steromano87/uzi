package messaging

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"time"
)

type Message struct {
	ID        string          `json:"id"`
	AnswersTo string          `json:"answersTo"`
	Timestamp time.Time       `json:"timestamp"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

func NewMessage(msgType string, payload any) (Message, error) {
	id, _ := uuid.NewRandom()
	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return Message{}, err
	}

	message := Message{
		ID:        id.String(),
		Timestamp: time.Now(),
		Type:      msgType,
		Payload:   marshalledPayload,
	}

	return message, nil
}

func NewAnswerMessage(msgType string, originalMsgID string, payload any) (Message, error) {
	id, _ := uuid.NewRandom()
	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return Message{}, err
	}

	message := Message{
		ID:        id.String(),
		AnswersTo: originalMsgID,
		Timestamp: time.Now(),
		Type:      msgType,
		Payload:   marshalledPayload,
	}

	return message, nil
}

func NewAcknowledgeMessage(originalMsgID string) Message {
	message, _ := NewAnswerMessage(AcknowledgeMsgType, originalMsgID, nil)
	return message
}

func NewPingMessage() Message {
	message, _ := NewMessage(PingMsgType, nil)
	return message
}

func NewPongMessage(pingMsgID string) Message {
	message, _ := NewAnswerMessage(PongMsgType, pingMsgID, nil)
	return message
}

func (m Message) DecodePayload() (any, error) {
	rawPayload := m.Payload

	switch m.Type {
	case SampleMsgType:
		var payload []db.Sample
		err := json.Unmarshal(rawPayload, &payload)
		return payload, err

	case MostMetricsMsgType:
		var payload db.HostMetric
		err := json.Unmarshal(rawPayload, &payload)
		return payload, err

	case EventMsgType:
		var payload db.Event
		err := json.Unmarshal(rawPayload, &payload)
		return payload, err

	case LogMsgType:
		var payload []db.Log
		err := json.Unmarshal(rawPayload, &payload)
		return payload, err

	default:
		return nil, errors.New("unknown payload type " + m.Type)
	}
}
