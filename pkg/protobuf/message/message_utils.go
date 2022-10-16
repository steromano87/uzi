package message

import (
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

func NewAcknowledgeEnvelope(answersTo string, ok bool, details *string) *Envelope {
	ack := &Envelope_Acknowledge{
		Acknowledge: &Acknowledge{
			Ok:      ok,
			Details: details,
		},
	}

	return NewResponseEnvelope(answersTo, ack)
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
