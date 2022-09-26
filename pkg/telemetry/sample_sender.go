package telemetry

import (
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"sync"
)

type SampleSender struct {
	messenger  messaging.Messenger
	bufferSize int

	queuedSamples []db.Sample
	mu            sync.Mutex
}

func NewSampleSender(messenger messaging.Messenger, bufferSize int) *SampleSender {
	sender := new(SampleSender)
	sender.messenger = messenger
	sender.bufferSize = bufferSize
	sender.queuedSamples = make([]db.Sample, 0)

	return sender
}

func (s *SampleSender) Collect(sample db.Sample) {
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
	sampleMessage, err := messaging.NewMessage(messaging.SampleMsgType, s.queuedSamples)
	if err != nil {
		// TODO: add error remote logging
		return
	}

	s.messenger.Send(sampleMessage)
	s.queuedSamples = make([]db.Sample, 0)
}
