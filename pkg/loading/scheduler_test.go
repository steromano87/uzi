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
	profile   loading.Profiler
	scheduler loading.Scheduler
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
}

func (s *SchedulerTestSuite) TestSingleInjectorQuota() {
	injectorReferences := []injector.RemoteReference{
		{
			ID:        "first",
			Weight:    1,
			Messenger: &MockedMessenger{},
		},
	}
	s.scheduler.Injectors = injectorReferences
	err := s.scheduler.Schedule(1)

	if assert.NoError(s.T(), err) {
		messenger := injectorReferences[0].Messenger.(*MockedMessenger)
		assert.Len(s.T(), messenger.MessagesFromWrite, 1)

		message := messenger.MessagesFromWrite[0]
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 10}, message.Payload)
	}
}

func (s *SchedulerTestSuite) TestTwoInjectorsWithSameWeight() {
	injectorReferences := []injector.RemoteReference{
		{
			ID:        "first",
			Weight:    1,
			Messenger: &MockedMessenger{},
		},
		{
			ID:        "second",
			Weight:    1,
			Messenger: &MockedMessenger{},
		},
	}
	s.scheduler.Injectors = injectorReferences
	err := s.scheduler.Schedule(1)

	if assert.NoError(s.T(), err) {
		firstMessenger := injectorReferences[0].Messenger.(*MockedMessenger)
		assert.Len(s.T(), firstMessenger.MessagesFromWrite, 1)

		firstMessage := firstMessenger.MessagesFromWrite[0]
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 5}, firstMessage.Payload)

		secondMessenger := injectorReferences[1].Messenger.(*MockedMessenger)
		assert.Len(s.T(), secondMessenger.MessagesFromWrite, 1)

		secondMessage := secondMessenger.MessagesFromWrite[0]
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 5}, secondMessage.Payload)
	}
}

func (s *SchedulerTestSuite) TestTwoInjectorsWithDifferentWeight() {
	injectorReferences := []injector.RemoteReference{
		{
			ID:        "first",
			Weight:    8,
			Messenger: &MockedMessenger{},
		},
		{
			ID:        "second",
			Weight:    2,
			Messenger: &MockedMessenger{},
		},
	}
	s.scheduler.Injectors = injectorReferences
	err := s.scheduler.Schedule(1)

	if assert.NoError(s.T(), err) {
		firstMessenger := injectorReferences[0].Messenger.(*MockedMessenger)
		assert.Len(s.T(), firstMessenger.MessagesFromWrite, 1)

		firstMessage := firstMessenger.MessagesFromWrite[0]
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 8}, firstMessage.Payload)

		secondMessenger := injectorReferences[1].Messenger.(*MockedMessenger)
		assert.Len(s.T(), secondMessenger.MessagesFromWrite, 1)

		secondMessage := secondMessenger.MessagesFromWrite[0]
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 2}, secondMessage.Payload)
	}
}

func (s *SchedulerTestSuite) TestThreeInjectorsWithDifferentWeight() {
	injectorReferences := []injector.RemoteReference{
		{
			ID:        "first",
			Weight:    8,
			Messenger: &MockedMessenger{},
		},
		{
			ID:        "second",
			Weight:    2,
			Messenger: &MockedMessenger{},
		},
		{
			ID:        "third",
			Weight:    2,
			Messenger: &MockedMessenger{},
		},
	}
	s.scheduler.Injectors = injectorReferences
	err := s.scheduler.Schedule(1)

	if assert.NoError(s.T(), err) {
		firstMessenger := injectorReferences[0].Messenger.(*MockedMessenger)
		assert.Len(s.T(), firstMessenger.MessagesFromWrite, 1)

		firstMessage := firstMessenger.MessagesFromWrite[0]
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 6}, firstMessage.Payload)

		secondMessenger := injectorReferences[1].Messenger.(*MockedMessenger)
		assert.Len(s.T(), secondMessenger.MessagesFromWrite, 1)

		secondMessage := secondMessenger.MessagesFromWrite[0]
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 2}, secondMessage.Payload)

		thirdMessenger := injectorReferences[2].Messenger.(*MockedMessenger)
		assert.Len(s.T(), thirdMessenger.MessagesFromWrite, 1)

		thirdMessage := thirdMessenger.MessagesFromWrite[0]
		assert.Equal(s.T(), &messaging.RunnersQuotaUpdatePayload{Quota: 2}, thirdMessage.Payload)
	}
}

func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
