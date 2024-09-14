package telemetry_test

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TestHostMetricsStorer struct {
	Samples []*telemetry.HostMetrics
}

func (t *TestHostMetricsStorer) StoreHostMetrics(hostMetrics *telemetry.HostMetrics) error {
	t.Samples = append(t.Samples, hostMetrics)
	return nil
}

type HostMetricsProbeTestSuite struct {
	suite.Suite
	storer *TestHostMetricsStorer
}

func (s *HostMetricsProbeTestSuite) SetupTest() {
	s.storer = &TestHostMetricsStorer{
		Samples: make([]*telemetry.HostMetrics, 0),
	}
}

func (s *HostMetricsProbeTestSuite) TestNewHostMetricsProbe() {
	collector := telemetry.NewHostMetricsProbe(s.storer)
	assert.IsType(s.T(), &telemetry.HostMetricsProbe{}, collector)
}

func (s *HostMetricsProbeTestSuite) TestMetricsCollection() {
	pollingInterval := 250 * time.Millisecond
	measuringInterval := 50 * time.Millisecond
	collector := telemetry.NewHostMetricsProbe(s.storer)
	ctx, cancelFunc := context.WithCancel(context.TODO())

	go func() {
		_ = collector.Serve(ctx, pollingInterval, measuringInterval)
	}()

	time.Sleep(2 * pollingInterval)
	cancelFunc()

	if assert.NotEmpty(s.T(), s.storer.Samples) {
		metric := s.storer.Samples[0]

		if assert.IsType(s.T(), &telemetry.HostMetrics{}, metric) {
			assert.GreaterOrEqual(s.T(), metric.GetCpu(), 0.0)
			assert.Greater(s.T(), metric.GetMemory().GetTotal(), uint64(0))
			assert.LessOrEqual(s.T(), metric.GetMemory().GetUsed(), metric.GetMemory().GetTotal())
			assert.Greater(s.T(), metric.GetStorage().GetTotal(), uint64(0))
			assert.LessOrEqual(s.T(), metric.GetStorage().GetUsed(), metric.GetStorage().GetTotal())
		}
	}
}

func TestHostMetricsCollectorTestSuite(t *testing.T) {
	suite.Run(t, new(HostMetricsProbeTestSuite))
}
