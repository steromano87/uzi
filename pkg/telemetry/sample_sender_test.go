package telemetry_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/protobuf/message"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SampleSenderTestSuite struct {
	suite.Suite
	bossMessenger   messaging.Messenger
	minionMessenger messaging.Messenger
}

func (s *SampleSenderTestSuite) SetupTest() {
	s.bossMessenger, s.minionMessenger = messaging.NewChannelMessengerPair(100)
}

func (s *SampleSenderTestSuite) TestNewSampleSender() {
	sender := telemetry.NewSampleSender(s.minionMessenger, 10)

	assert.IsType(s.T(), &telemetry.SampleSender{}, sender)
}

func (s *SampleSenderTestSuite) TestCollectBelowBufferLimit() {
	sample := &message.Sample{}

	sender := telemetry.NewSampleSender(s.minionMessenger, 10)
	sender.Collect(sample)
	sender.Collect(sample)

	// Check that no message was actually sent
	var noValue bool

	select {
	case <-s.bossMessenger.Receive():
		noValue = false

	default:
		noValue = true
	}

	assert.True(s.T(), noValue)
}

func (s *SampleSenderTestSuite) TestSampleSendingWithManualFlush() {
	sample := &message.Sample{}

	sender := telemetry.NewSampleSender(s.minionMessenger, 10)
	sender.Collect(sample)
	sender.Collect(sample)

	sender.Flush()

	var msg *message.Envelope

	select {
	case msg = <-s.bossMessenger.Receive():

	default:
		assert.Fail(s.T(), "no msg was sent")
	}

	if assert.IsType(s.T(), &message.Envelope{}, msg) {
		payload := msg.GetSamples()
		if assert.NotNil(s.T(), payload) {
			assert.IsType(s.T(), []*message.Sample{}, payload.GetSample())
			assert.Len(s.T(), payload.GetSample(), 2)
		}
	}
}

func (s *SampleSenderTestSuite) TestSampleSendingWithAutomaticFlush() {
	sample := &message.Sample{}

	sender := telemetry.NewSampleSender(s.minionMessenger, 2)
	sender.Collect(sample)
	sender.Collect(sample)

	var msg *message.Envelope

	select {
	case msg = <-s.bossMessenger.Receive():

	default:
		assert.Fail(s.T(), "no msg was sent")
	}

	if assert.IsType(s.T(), &message.Envelope{}, msg) {
		payload := msg.GetSamples()
		if assert.NotNil(s.T(), payload) {
			assert.IsType(s.T(), []*message.Sample{}, payload.GetSample())
			assert.Len(s.T(), payload.GetSample(), 2)
		}
	}
}

func TestSampleSenderSuite(t *testing.T) {
	suite.Run(t, new(SampleSenderTestSuite))
}
