package telemetry_test

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TestHostMetricsSaver struct {
	Samples []*telemetry.HostMetricsSample
}

func (t *TestHostMetricsSaver) SaveHostMetricsSample(hostMetrics *telemetry.HostMetricsSample) {
	t.Samples = append(t.Samples, hostMetrics)
}

type HostMetricsCollectorTestSuite struct {
	suite.Suite
}

func (s *HostMetricsCollectorTestSuite) TestNewHostMetricsCollector() {
	collector := telemetry.NewHostMetricsCollector(1*time.Second, 200*time.Millisecond)
	assert.IsType(s.T(), &telemetry.HostMetricsCollector{}, collector)
}

func (s *HostMetricsCollectorTestSuite) TestMetricsCollection() {
	pollingInterval := 250 * time.Millisecond
	measuringInterval := 50 * time.Millisecond
	collector := telemetry.NewHostMetricsCollector(pollingInterval, measuringInterval)
	ctx, cancelFunc := context.WithCancel(context.TODO())
	testSaver := &TestHostMetricsSaver{
		Samples: make([]*telemetry.HostMetricsSample, 0),
	}

	collector.Start(ctx, testSaver)

	time.Sleep(2 * pollingInterval)
	cancelFunc()

	if assert.NotEmpty(s.T(), testSaver.Samples) {
		metric := testSaver.Samples[0]

		if assert.IsType(s.T(), &telemetry.HostMetricsSample{}, metric) {
			assert.GreaterOrEqual(s.T(), metric.GetCpu(), 0.0)
			assert.Greater(s.T(), metric.GetMemory().GetTotal(), uint64(0))
			assert.LessOrEqual(s.T(), metric.GetMemory().GetUsed(), metric.GetMemory().GetTotal())
			assert.Greater(s.T(), metric.GetStorage().GetTotal(), uint64(0))
			assert.LessOrEqual(s.T(), metric.GetStorage().GetUsed(), metric.GetStorage().GetTotal())
		}
	}
}

func TestHostMetricsCollectorTestSuite(t *testing.T) {
	suite.Run(t, new(HostMetricsCollectorTestSuite))
}
