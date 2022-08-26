package loading_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

///////////////////////////////////////////////////

type MockedMessenger struct {
	MessageForRead    messaging.Message
	MessagesFromWrite []messaging.Message
}

func (m *MockedMessenger) Receive() (messaging.Message, error) {
	return m.MessageForRead, nil
}

func (m *MockedMessenger) Send(message messaging.Message) error {
	m.MessagesFromWrite = append(m.MessagesFromWrite, message)
	return nil
}

func (m *MockedMessenger) SendPing() error {
	panic("not required")
}

func (m *MockedMessenger) SendPong(_ string) error {
	panic("not required")
}

func (m *MockedMessenger) Close() {
	panic("not required")
}

///////////////////////////////////////////////////

type SchedulerTestSuite struct {
	suite.Suite
	profile         loading.Profiler
	scheduler       loading.Scheduler
	bossMessenger   messaging.Messenger
	minionMessenger messaging.Messenger
}

func (s *SchedulerTestSuite) SetupTest() {
	s.profile = loading.LinearRamp{
		Shooters:     10,
		InitialDelay: 0,
		RampUp:       0,
		Sustain:      1000,
		RampDown:     0,
	}
	s.scheduler = loading.Scheduler{
		Profile:   s.profile,
		Injectors: nil,
	}

	s.bossMessenger, s.minionMessenger = messaging.NewChannelMessengerPair(100)
}

func (s *SchedulerTestSuite) TestSingleInjectorQuota() {
	injectorReferences := []injector.RemoteReference{
		{
			ID:        "first",
			Weight:    1,
			Messenger: s.bossMessenger,
		},
	}
	s.scheduler.Injectors = injectorReferences
	s.scheduler.Schedule(1)
	message, ok := <-s.minionMessenger.Receive()

	if assert.True(s.T(), ok) {
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 10}, message.Payload)
	}
}

func (s *SchedulerTestSuite) TestTwoInjectorsWithSameWeight() {
	injectorReferences := []injector.RemoteReference{
		{
			ID:        "first",
			Weight:    1,
			Messenger: s.bossMessenger,
		},
		{
			ID:        "second",
			Weight:    1,
			Messenger: s.bossMessenger,
		},
	}
	s.scheduler.Injectors = injectorReferences
	s.scheduler.Schedule(1)
	firstMessage := <-s.minionMessenger.Receive()
	secondMessage, ok := <-s.minionMessenger.Receive()

	if assert.True(s.T(), ok) {
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 5}, firstMessage.Payload)
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 5}, secondMessage.Payload)
	}
}

func (s *SchedulerTestSuite) TestTwoInjectorsWithDifferentWeight() {
	injectorReferences := []injector.RemoteReference{
		{
			ID:        "first",
			Weight:    8,
			Messenger: s.bossMessenger,
		},
		{
			ID:        "second",
			Weight:    2,
			Messenger: s.bossMessenger,
		},
	}
	s.scheduler.Injectors = injectorReferences
	s.scheduler.Schedule(1)
	firstMessage := <-s.minionMessenger.Receive()
	secondMessage, ok := <-s.minionMessenger.Receive()

	if assert.True(s.T(), ok) {
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 8}, firstMessage.Payload)
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 2}, secondMessage.Payload)
	}
}

func (s *SchedulerTestSuite) TestThreeInjectorsWithDifferentWeight() {
	injectorReferences := []injector.RemoteReference{
		{
			ID:        "first",
			Weight:    8,
			Messenger: s.bossMessenger,
		},
		{
			ID:        "second",
			Weight:    2,
			Messenger: s.bossMessenger,
		},
		{
			ID:        "third",
			Weight:    2,
			Messenger: s.bossMessenger,
		},
	}
	s.scheduler.Injectors = injectorReferences
	s.scheduler.Schedule(1)
	firstMessage := <-s.minionMessenger.Receive()
	secondMessage := <-s.minionMessenger.Receive()
	thirdMessage, ok := <-s.minionMessenger.Receive()

	if assert.True(s.T(), ok) {
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 6}, firstMessage.Payload)
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 2}, secondMessage.Payload)
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 2}, thirdMessage.Payload)
	}
}

func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
