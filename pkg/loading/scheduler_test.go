package loading_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

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

type SchedulerTestSuite struct {
	suite.Suite
	messenger MockedMessenger
	profile   loading.Profiler
	scheduler loading.Scheduler
}

func (s *SchedulerTestSuite) SetupTest() {
	s.messenger = MockedMessenger{}
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
	injectorReferences := []loading.InjectorRemoteReference{
		{
			Name:      "first",
			Weight:    1,
			Messenger: &s.messenger,
		},
	}
	s.scheduler.Injectors = injectorReferences
	err := s.scheduler.Schedule(1)

	if assert.NoError(s.T(), err) {
		assert.Len(s.T(), s.messenger.MessagesFromWrite, 1)

		message := s.messenger.MessagesFromWrite[0]
		expectedPayload := map[string]any{
			"injectorID": "first",
			"newQuota":   10,
		}
		assert.Equal(s.T(), expectedPayload, message.Payload)
	}
}

func (s *SchedulerTestSuite) TestTwoInjectorsWithSameWeight() {
	injectorReferences := []loading.InjectorRemoteReference{
		{
			Name:      "first",
			Weight:    1,
			Messenger: &s.messenger,
		},
		{
			Name:      "second",
			Weight:    1,
			Messenger: &s.messenger,
		},
	}
	s.scheduler.Injectors = injectorReferences
	err := s.scheduler.Schedule(1)

	if assert.NoError(s.T(), err) {
		assert.Len(s.T(), s.messenger.MessagesFromWrite, 2)

		for _, message := range s.messenger.MessagesFromWrite {
			payload := message.Payload
			injectorID := payload["injectorID"].(string)
			if injectorID == "first" {
				expectedPayload := map[string]any{
					"injectorID": "first",
					"newQuota":   5,
				}
				assert.Equal(s.T(), expectedPayload, payload)
			} else {
				expectedPayload := map[string]any{
					"injectorID": "second",
					"newQuota":   5,
				}
				assert.Equal(s.T(), expectedPayload, payload)
			}
		}
	}
}

func (s *SchedulerTestSuite) TestTwoInjectorsWithDifferentWeight() {
	injectorReferences := []loading.InjectorRemoteReference{
		{
			Name:      "first",
			Weight:    8,
			Messenger: &s.messenger,
		},
		{
			Name:      "second",
			Weight:    2,
			Messenger: &s.messenger,
		},
	}
	s.scheduler.Injectors = injectorReferences
	err := s.scheduler.Schedule(1)

	if assert.NoError(s.T(), err) {
		assert.Len(s.T(), s.messenger.MessagesFromWrite, 2)

		for _, message := range s.messenger.MessagesFromWrite {
			payload := message.Payload
			injectorID := payload["injectorID"].(string)
			if injectorID == "first" {
				expectedPayload := map[string]any{
					"injectorID": "first",
					"newQuota":   8,
				}
				assert.Equal(s.T(), expectedPayload, payload)
			} else {
				expectedPayload := map[string]any{
					"injectorID": "second",
					"newQuota":   2,
				}
				assert.Equal(s.T(), expectedPayload, payload)
			}
		}
	}
}

func (s *SchedulerTestSuite) TestThreeInjectorsWithDifferentWeight() {
	injectorReferences := []loading.InjectorRemoteReference{
		{
			Name:      "first",
			Weight:    8,
			Messenger: &s.messenger,
		},
		{
			Name:      "second",
			Weight:    2,
			Messenger: &s.messenger,
		},
		{
			Name:      "third",
			Weight:    2,
			Messenger: &s.messenger,
		},
	}
	s.scheduler.Injectors = injectorReferences
	err := s.scheduler.Schedule(1)

	if assert.NoError(s.T(), err) {
		assert.Len(s.T(), s.messenger.MessagesFromWrite, 3)

		for _, message := range s.messenger.MessagesFromWrite {
			payload := message.Payload
			injectorID := payload["injectorID"].(string)
			if injectorID == "first" {
				expectedPayload := map[string]any{
					"injectorID": "first",
					"newQuota":   6,
				}
				assert.Equal(s.T(), expectedPayload, payload)
			} else if injectorID == "second" {
				expectedPayload := map[string]any{
					"injectorID": "second",
					"newQuota":   2,
				}
				assert.Equal(s.T(), expectedPayload, payload)
			} else {
				expectedPayload := map[string]any{
					"injectorID": "third",
					"newQuota":   2,
				}
				assert.Equal(s.T(), expectedPayload, payload)
			}
		}
	}
}

func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
