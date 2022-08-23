package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/steromano87/harkonnen/v1/pkg/model"
)

const SampleMsgId = "SAMPLE"

type SamplePayload struct {
	Samples []model.Sample `json:"samples"`
}

func NewSampleMessage(samples []model.Sample) Message {
	return NewRawMessage(RemoteLogMsgId, &SamplePayload{Samples: samples})
}

func (s *SamplePayload) UnmarshalJSON(bytes []byte) error {
	var intermediate map[string]any

	if err := json.Unmarshal(bytes, &intermediate); err != nil {
		return err
	}

	samples, ok := intermediate[SampleMsgId].([]model.Sample)
	if !ok {
		return errors.New(fmt.Sprintf("error parsing samples '%s' field: %x", SampleMsgId, samples))
	}

	s.Samples = samples
	return nil
}
