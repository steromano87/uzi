package telemetry_test

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
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
		if assert.Equal(s.T(), messaging.HostMetricsMsgId, message.Type) {
			payload := message.Payload

			if assert.IsType(s.T(), &messaging.HostMetricsPayload{}, payload) {
				castPayload := payload.(*messaging.HostMetricsPayload)

				assert.Greater(s.T(), castPayload.CPU, 0.0)
				assert.Greater(s.T(), castPayload.Memory.Total, uint64(0))
				assert.LessOrEqual(s.T(), castPayload.Memory.Used, castPayload.Memory.Total)
				assert.Greater(s.T(), castPayload.Disk.Total, uint64(0))
				assert.LessOrEqual(s.T(), castPayload.Disk.Used, castPayload.Disk.Total)
			}
		}

	default:
		assert.Fail(s.T(), "no metrics message was sent")
	}
}

func TestHostMetricsCollectorTestSuite(t *testing.T) {
	suite.Run(t, new(HostMetricsCollectorTestSuite))
}
