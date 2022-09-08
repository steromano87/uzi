package injector

import (
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"sync"
)

type SampleSender struct {
	messenger  messaging.Messenger
	bufferSize int

	queuedSamples []model.Sample
	mu            sync.Mutex
}

func NewSampleSender(messenger messaging.Messenger, bufferSize int) *SampleSender {
	sender := new(SampleSender)
	sender.messenger = messenger
	sender.bufferSize = bufferSize
	sender.queuedSamples = make([]model.Sample, 0)

	return sender
}

func (s *SampleSender) Collect(sample model.Sample) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.queuedSamples = append(s.queuedSamples, sample)
	if len(s.queuedSamples) >= s.bufferSize {
		s.Flush()
	}
}

func (s *SampleSender) Flush() {
	sampleMessage := messaging.NewSampleMessage(s.queuedSamples)
	s.messenger.Send(sampleMessage)
	s.queuedSamples = make([]model.Sample, 0)
}
