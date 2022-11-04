package telemetry

import (
	"sync"
)

type SampleCollector struct {
	samples []*Sample
	mu      sync.Mutex
}

func NewSampleCollector() *SampleCollector {
	collector := new(SampleCollector)
	collector.samples = make([]*Sample, 0)

	return collector
}

func (s *SampleCollector) AddSample(sample *Sample) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.samples = append(s.samples, sample)
}

func (s *SampleCollector) GetSamples() []*Sample {
	s.mu.Lock()
	defer s.mu.Unlock()

	var outputSamples []*Sample
	copy(outputSamples, s.samples)
	s.samples = make([]*Sample, 0)
	return outputSamples
}
