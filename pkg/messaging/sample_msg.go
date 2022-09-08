package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/steromano87/harkonnen/v1/pkg/db"
)

const SampleMsgId = "SAMPLE"

type SamplePayload struct {
	Samples []db.Sample `json:"samples"`
}

func NewSampleMessage(samples []db.Sample) Message {
	return NewRawMessage(RemoteLogMsgId, &SamplePayload{Samples: samples})
}

func (s *SamplePayload) UnmarshalJSON(bytes []byte) error {
	var intermediate map[string]any

	if err := json.Unmarshal(bytes, &intermediate); err != nil {
		return err
	}

	samples, ok := intermediate[SampleMsgId].([]db.Sample)
	if !ok {
		return errors.New(fmt.Sprintf("error parsing samples '%s' field: %x", SampleMsgId, samples))
	}

	s.Samples = samples
	return nil
}
