package messaging

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	ID        string         `json:"id"`
	AnswersTo string         `json:"answersTo"`
	Timestamp time.Time      `json:"timestamp"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload"`
}

func NewMessage(msgType string, payload map[string]any) Message {
	ID, _ := uuid.NewRandom()
	return Message{
		ID:        ID.String(),
		Timestamp: time.Now(),
		Type:      msgType,
		Payload:   payload,
	}
}

func NewAnswerMessage(msgType string, originalMsgID string, payload map[string]any) Message {
	ID, _ := uuid.NewRandom()
	return Message{
		ID:        ID.String(),
		AnswersTo: originalMsgID,
		Timestamp: time.Now(),
		Type:      msgType,
		Payload:   payload,
	}
}
