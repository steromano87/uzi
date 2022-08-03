package loading_test

import (
	"github.com/gorilla/websocket"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"github.com/steromano87/harkonnen/v1/pkg/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MockedMessenger struct {
	MessageTypeForRead int
	MessageForRead     []byte

	MessagesTypeFromWrite []int
	MessagesFromWrite     [][]byte
}

func (m *MockedMessenger) Read() (int, []byte, error) {
	return m.MessageTypeForRead, m.MessageForRead, nil
}

func (m *MockedMessenger) Write(msgType int, payload []byte) error {
	m.MessagesTypeFromWrite = append(m.MessagesTypeFromWrite, msgType)
	m.MessagesFromWrite = append(m.MessagesFromWrite, payload)

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
		assert.Len(s.T(), s.messenger.MessagesTypeFromWrite, 1)
		assert.Len(s.T(), s.messenger.MessagesFromWrite, 1)

		msgType := s.messenger.MessagesTypeFromWrite[0]
		assert.Equal(s.T(), websocket.TextMessage, msgType)

		rawMessage := s.messenger.MessagesFromWrite[0]
		message, err := network.UnpackMessage(rawMessage)

		if assert.NoError(s.T(), err) {
			// Using float64 instead of int because JSON only knows floats
			// (see https://pkg.go.dev/encoding/json#Unmarshal)
			expectedPayload := map[string]any{
				"injectorID": "first",
				"newQuota":   10.0,
			}
			assert.Equal(s.T(), expectedPayload, message.Payload)
		}
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
		assert.Len(s.T(), s.messenger.MessagesTypeFromWrite, 2)
		assert.Len(s.T(), s.messenger.MessagesFromWrite, 2)

		for _, msgType := range s.messenger.MessagesTypeFromWrite {
			assert.Equal(s.T(), websocket.TextMessage, msgType)
		}

		for _, rawMessage := range s.messenger.MessagesFromWrite {
			message, err := network.UnpackMessage(rawMessage)

			if assert.NoError(s.T(), err) {
				payloadMap, _ := message.PayloadMap()
				injectorID := payloadMap["injectorID"].(string)
				if injectorID == "first" {
					expectedPayload := map[string]any{
						"injectorID": "first",
						"newQuota":   5.0,
					}
					assert.Equal(s.T(), expectedPayload, payloadMap)
				} else {
					expectedPayload := map[string]any{
						"injectorID": "second",
						"newQuota":   5.0,
					}
					assert.Equal(s.T(), expectedPayload, payloadMap)
				}
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
		assert.Len(s.T(), s.messenger.MessagesTypeFromWrite, 2)
		assert.Len(s.T(), s.messenger.MessagesFromWrite, 2)

		for _, msgType := range s.messenger.MessagesTypeFromWrite {
			assert.Equal(s.T(), websocket.TextMessage, msgType)
		}

		for _, rawMessage := range s.messenger.MessagesFromWrite {
			message, err := network.UnpackMessage(rawMessage)

			if assert.NoError(s.T(), err) {
				payloadMap, _ := message.PayloadMap()
				injectorID := payloadMap["injectorID"].(string)
				if injectorID == "first" {
					expectedPayload := map[string]any{
						"injectorID": "first",
						"newQuota":   8.0,
					}
					assert.Equal(s.T(), expectedPayload, payloadMap)
				} else {
					expectedPayload := map[string]any{
						"injectorID": "second",
						"newQuota":   2.0,
					}
					assert.Equal(s.T(), expectedPayload, payloadMap)
				}
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
		assert.Len(s.T(), s.messenger.MessagesTypeFromWrite, 3)
		assert.Len(s.T(), s.messenger.MessagesFromWrite, 3)

		for _, msgType := range s.messenger.MessagesTypeFromWrite {
			assert.Equal(s.T(), websocket.TextMessage, msgType)
		}

		for _, rawMessage := range s.messenger.MessagesFromWrite {
			message, err := network.UnpackMessage(rawMessage)

			if assert.NoError(s.T(), err) {
				payloadMap, _ := message.PayloadMap()
				injectorID := payloadMap["injectorID"].(string)
				if injectorID == "first" {
					expectedPayload := map[string]any{
						"injectorID": "first",
						"newQuota":   6.0,
					}
					assert.Equal(s.T(), expectedPayload, payloadMap)
				} else if injectorID == "second" {
					expectedPayload := map[string]any{
						"injectorID": "second",
						"newQuota":   2.0,
					}
					assert.Equal(s.T(), expectedPayload, payloadMap)
				} else {
					expectedPayload := map[string]any{
						"injectorID": "third",
						"newQuota":   2.0,
					}
					assert.Equal(s.T(), expectedPayload, payloadMap)
				}
			}
		}
	}
}

func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
