package protobuf

import (
	"github.com/google/uuid"
	"github.com/steromano87/harkonnen/v1/pkg/protobuf/message"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewEnvelope(payload message.Payload) message.Envelope {
	return message.Envelope{
		Id:        uuid.NewString(),
		Timestamp: timestamppb.Now(),
		AnswersTo: nil,
		Payload:   payload,
	}
}

func NewResponseEnvelope(answersTo string, payload message.Payload) message.Envelope {
	return message.Envelope{
		Id:        uuid.NewString(),
		Timestamp: timestamppb.Now(),
		AnswersTo: &answersTo,
		Payload:   payload,
	}
}
