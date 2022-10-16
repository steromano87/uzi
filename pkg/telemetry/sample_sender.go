package telemetry

import (
	"github.com/steromano87/harkonnen/v1/pkg/message"
	"sync"
)

type SampleSender struct {
	messenger  message.Bridge
	bufferSize int

	queuedSamples []*message.Sample
	mu            sync.Mutex
}

func NewSampleSender(messenger message.Bridge, bufferSize int) *SampleSender {
	sender := new(SampleSender)
	sender.messenger = messenger
	sender.bufferSize = bufferSize
	sender.queuedSamples = make([]*message.Sample, 0)

	return sender
}

func (s *SampleSender) Collect(sample *message.Sample) {
	s.mu.Lock()
	s.queuedSamples = append(s.queuedSamples, sample)
	s.mu.Unlock()

	if len(s.queuedSamples) >= s.bufferSize {
		s.Flush()
	}
}

func (s *SampleSender) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	payload := message.Envelope_Samples{Samples: &message.Samples{Sample: s.queuedSamples}}
	msg := message.NewEnvelope(&payload)

	s.messenger.Send(msg)
	s.queuedSamples = make([]*message.Sample, 0)
}
