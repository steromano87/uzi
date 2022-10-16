package telemetry_test

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	message2 "github.com/steromano87/harkonnen/v1/pkg/protobuf/message"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type HostMetricsCollectorTestSuite struct {
	suite.Suite
	bossMessenger   messaging.Messenger
	minionMessenger messaging.Messenger
}

func (s *HostMetricsCollectorTestSuite) SetupTest() {
	s.bossMessenger, s.minionMessenger = messaging.NewChannelMessengerPair(100)
}

func (s *HostMetricsCollectorTestSuite) TestNewHostMetricsCollector() {
	collector := telemetry.NewHostMetricsSender(s.minionMessenger)
	assert.IsType(s.T(), &telemetry.HostMetricsSender{}, collector)
}

func (s *HostMetricsCollectorTestSuite) TestMetricsCollection() {
	collector := telemetry.NewHostMetricsSender(s.minionMessenger)
	ctx, cancelFunc := context.WithCancel(context.TODO())

	pollingInterval := 250 * time.Millisecond
	measuringInterval := 50 * time.Millisecond
	collector.Start(ctx, pollingInterval, measuringInterval)

	time.Sleep(2 * pollingInterval)
	cancelFunc()

	select {
	case message := <-s.bossMessenger.Receive():
		if assert.IsType(s.T(), &message2.Envelope_HostMetrics{}, message.Payload) {
			payload := message.GetHostMetrics()
			if assert.NotNil(s.T(), payload) {
				if assert.IsType(s.T(), &message2.HostMetrics{}, payload) {
					assert.GreaterOrEqual(s.T(), payload.GetCpu(), 0.0)
					assert.Greater(s.T(), payload.GetMemory().GetTotal(), uint64(0))
					assert.LessOrEqual(s.T(), payload.GetMemory().GetUsed(), payload.GetMemory().GetTotal())
					assert.Greater(s.T(), payload.GetStorage().GetTotal(), uint64(0))
					assert.LessOrEqual(s.T(), payload.GetStorage().GetUsed(), payload.GetStorage().GetTotal())
				}
			}
		}

	default:
		assert.Fail(s.T(), "no metrics message was sent")
	}
}

func TestHostMetricsCollectorTestSuite(t *testing.T) {
	suite.Run(t, new(HostMetricsCollectorTestSuite))
}
