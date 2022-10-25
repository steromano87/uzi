package message

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/steromano87/harkonnen/v1/pkg/utils"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func NewEnvelope(payload Payload) *Envelope {
	return &Envelope{
		Id:        uuid.NewString(),
		Timestamp: timestamppb.Now(),
		AnswersTo: nil,
		Payload:   payload,
	}
}

func NewResponseEnvelope(answersTo string, payload Payload) *Envelope {
	return &Envelope{
		Id:        uuid.NewString(),
		Timestamp: timestamppb.Now(),
		AnswersTo: &answersTo,
		Payload:   payload,
	}
}

func NewHelloEnvelope(harkonnenVersion string) *Envelope {
	hello := &Envelope_Hello{
		Hello: &Hello{
			HarkonnenVersion: harkonnenVersion,
		},
	}

	return NewEnvelope(hello)
}

func NewAcknowledgeEnvelope(answersTo string, ok bool, details *string, newStatus *string) *Envelope {
	ack := &Envelope_Acknowledge{
		Acknowledge: &Acknowledge{
			Ok:      ok,
			Details: details,
			Status:  newStatus,
		},
	}

	return NewResponseEnvelope(answersTo, ack)
}

func NewPositiveAcknowledgeEnvelope(answersTo string) *Envelope {
	return NewAcknowledgeEnvelope(answersTo, true, nil, nil)
}

func NewStatusChangeAcknowledgeEnvelope(answersTo string, newStatus string) *Envelope {
	return NewAcknowledgeEnvelope(answersTo, true, nil, &newStatus)
}

func NewErrorAcknowledgeEnvelope(answersTo string, errorDescription string, err error) *Envelope {
	var details string
	if err != nil {
		details = fmt.Sprintf("%s: %s", errorDescription, err)
	} else {
		details = errorDescription
	}
	return NewAcknowledgeEnvelope(answersTo, false, &details, nil)
}

func NewForcedShutdownRequestEnvelope() *Envelope {
	return NewEnvelope(&Envelope_ForcedShutdownRequest{
		ForcedShutdownRequest: &ForcedShutdownRequest{},
	})
}

func NewGracefulShutdownRequestEnvelope(timeout time.Duration) *Envelope {
	return NewEnvelope(&Envelope_GracefulShutdownRequest{
		GracefulShutdownRequest: &GracefulShutdownRequest{
			Timeout: durationpb.New(timeout),
		},
	})
}

func NewWorkingFolderInitEnvelope(workingFolder string) (*Envelope, error) {
	compressedFolder, err := utils.ZipFolder(workingFolder)
	if err != nil {
		return nil, err
	}

	return NewEnvelope(&Envelope_WorkingFolderInit{
		WorkingFolderInit: &WorkingFolderInit{
			CompressedWorkingFolder: compressedFolder,
			CompressionAlgorithm:    CompressionAlgorithm_ZIP,
		},
	}), nil
}
