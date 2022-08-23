package messaging

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	ID        string    `json:"id"`
	AnswersTo string    `json:"answersTo"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Payload   Payload   `json:"payload,omitempty"`
}

func NewRawMessage(msgType string, payload Payload) Message {
	id, _ := uuid.NewRandom()
	message := Message{
		ID:        id.String(),
		Timestamp: time.Now(),
		Type:      msgType,
		Payload:   payload,
	}

	return message
}

func NewRawAnswerMessage(msgType string, originalMsgID string, payload Payload) Message {
	id, _ := uuid.NewRandom()
	message := Message{
		ID:        id.String(),
		AnswersTo: originalMsgID,
		Timestamp: time.Now(),
		Type:      msgType,
		Payload:   payload,
	}

	return message
}
