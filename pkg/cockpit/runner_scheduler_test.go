package cockpit_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/cockpit"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SchedulerTestSuite struct {
	suite.Suite
	profile         cockpit.LoadProfile
	scheduler       cockpit.RunnerScheduler
	bossMessenger   messaging.Messenger
	minionMessenger messaging.Messenger
}

func (s *SchedulerTestSuite) SetupTest() {
	s.profile = cockpit.LinearRamp{
		Runners:      10,
		InitialDelay: 0,
		RampUp:       0,
		Sustain:      1000,
		RampDown:     0,
	}
	s.scheduler = cockpit.RunnerScheduler{
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
		payload, err := message.DecodePayload()

		if assert.NoError(s.T(), err) {
			if assert.IsType(s.T(), db.Event{}, payload) {
				assert.EqualValues(s.T(), 10, payload.(db.Event).Data["runnerQuota"])
			}
		}
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
		payload, err := firstMessage.DecodePayload()

		if assert.NoError(s.T(), err) {
			if assert.IsType(s.T(), db.Event{}, payload) {
				assert.EqualValues(s.T(), 5, payload.(db.Event).Data["runnerQuota"])
			}
		}

		payload, err = secondMessage.DecodePayload()

		if assert.NoError(s.T(), err) {
			if assert.IsType(s.T(), db.Event{}, payload) {
				assert.EqualValues(s.T(), 5, payload.(db.Event).Data["runnerQuota"])
			}
		}
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
		payload, err := firstMessage.DecodePayload()

		if assert.NoError(s.T(), err) {
			if assert.IsType(s.T(), db.Event{}, payload) {
				assert.EqualValues(s.T(), 8, payload.(db.Event).Data["runnerQuota"])
			}
		}

		payload, err = secondMessage.DecodePayload()

		if assert.NoError(s.T(), err) {
			if assert.IsType(s.T(), db.Event{}, payload) {
				assert.EqualValues(s.T(), 2, payload.(db.Event).Data["runnerQuota"])
			}
		}
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
		payload, err := firstMessage.DecodePayload()

		if assert.NoError(s.T(), err) {
			if assert.IsType(s.T(), db.Event{}, payload) {
				assert.EqualValues(s.T(), 6, payload.(db.Event).Data["runnerQuota"])
			}
		}

		payload, err = secondMessage.DecodePayload()

		if assert.NoError(s.T(), err) {
			if assert.IsType(s.T(), db.Event{}, payload) {
				assert.EqualValues(s.T(), 2, payload.(db.Event).Data["runnerQuota"])
			}
		}

		payload, err = thirdMessage.DecodePayload()

		if assert.NoError(s.T(), err) {
			if assert.IsType(s.T(), db.Event{}, payload) {
				assert.EqualValues(s.T(), 2, payload.(db.Event).Data["runnerQuota"])
			}
		}
	}
}

func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
