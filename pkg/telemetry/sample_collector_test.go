package telemetry_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SampleCollectorTestSuite struct {
	suite.Suite
}

func (s *SampleCollectorTestSuite) TestNewSampleSender() {
	sampleCollector := telemetry.NewSampleCollector()

	assert.IsType(s.T(), &telemetry.SampleCollector{}, sampleCollector)
}

func (s *SampleCollectorTestSuite) TestSampleAdditionAndFlush() {
	originalSample := &telemetry.Sample{}

	sampleCollector := telemetry.NewSampleCollector()
	sampleCollector.AddSample(originalSample)
	sampleCollector.AddSample(originalSample)

	samples := sampleCollector.GetSamples()

	if assert.Len(s.T(), samples, 2) {
		assert.Equal(s.T(), originalSample, samples[0])
		assert.Equal(s.T(), originalSample, samples[1])
	}
}

func TestSampleCollectorSuite(t *testing.T) {
	suite.Run(t, new(SampleCollectorTestSuite))
}
