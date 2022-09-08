package messaging_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
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
	sender := messaging.NewSampleSender(s.minionMessenger, 10)

	assert.IsType(s.T(), &messaging.SampleSender{}, sender)
}

func (s *SampleSenderTestSuite) TestCollectBelowBufferLimit() {
	sample := db.Sample{}

	sender := messaging.NewSampleSender(s.minionMessenger, 10)
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
	sample := db.Sample{}

	sender := messaging.NewSampleSender(s.minionMessenger, 10)
	sender.Collect(sample)
	sender.Collect(sample)

	sender.Flush()

	var message messaging.Message

	select {
	case message = <-s.bossMessenger.Receive():

	default:
		assert.Fail(s.T(), "no message was sent")
	}

	if assert.IsType(s.T(), messaging.Message{}, message) {
		payload := message.Payload
		if assert.IsType(s.T(), &messaging.SamplePayload{}, payload) {
			assert.Len(s.T(), payload.(*messaging.SamplePayload).Samples, 2)
		}
	}
}

func (s *SampleSenderTestSuite) TestSampleSendingWithAutomaticFlush() {
	sample := db.Sample{}

	sender := messaging.NewSampleSender(s.minionMessenger, 2)
	sender.Collect(sample)
	sender.Collect(sample)

	var message messaging.Message

	select {
	case message = <-s.bossMessenger.Receive():

	default:
		assert.Fail(s.T(), "no message was sent")
	}

	if assert.IsType(s.T(), messaging.Message{}, message) {
		payload := message.Payload
		if assert.IsType(s.T(), &messaging.SamplePayload{}, payload) {
			assert.Len(s.T(), payload.(*messaging.SamplePayload).Samples, 2)
		}
	}
}

func TestSampleSenderSuite(t *testing.T) {
	suite.Run(t, new(SampleSenderTestSuite))
}
