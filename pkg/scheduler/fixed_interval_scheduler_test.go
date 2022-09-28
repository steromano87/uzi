package scheduler_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SchedulerTestSuite struct {
	suite.Suite
	profile         scheduler.LoadProfile
	bossMessenger   messaging.Messenger
	minionMessenger messaging.Messenger
}

func (s *SchedulerTestSuite) SetupTest() {
	s.profile = scheduler.LinearRamp{
		Runners:      10,
		InitialDelay: 0,
		RampUp:       0,
		Sustain:      1000,
		RampDown:     0,
	}

	s.bossMessenger, s.minionMessenger = messaging.NewChannelMessengerPair(100)
}

func (s *SchedulerTestSuite) TestSingleInjectorQuota() {
	injectorReferences := map[string]injector.RemoteReference{
		"first": {
			Weight:    1,
			Messenger: s.bossMessenger,
		},
	}

	sched := scheduler.NewFixedIntervalScheduler(s.profile, &injectorReferences, 5)
	quotas := sched.At(1)

	assert.Equal(s.T(), 10, quotas["first"])
}

func (s *SchedulerTestSuite) TestTwoInjectorsWithSameWeight() {
	injectorReferences := map[string]injector.RemoteReference{
		"first": {
			Weight:    1,
			Messenger: s.bossMessenger,
		},
		"second": {
			Weight:    1,
			Messenger: s.bossMessenger,
		},
	}
	sched := scheduler.NewFixedIntervalScheduler(s.profile, &injectorReferences, 5)
	quotas := sched.At(1)

	assert.Equal(s.T(), 5, quotas["first"], "first weight is wrong")
	assert.Equal(s.T(), 5, quotas["second"], "second weight is wrong")
}

func (s *SchedulerTestSuite) TestTwoInjectorsWithDifferentWeight() {
	injectorReferences := map[string]injector.RemoteReference{
		"first": {
			Weight:    8,
			Messenger: s.bossMessenger,
		},
		"second": {
			Weight:    2,
			Messenger: s.bossMessenger,
		},
	}

	sched := scheduler.NewFixedIntervalScheduler(s.profile, &injectorReferences, 5)
	quotas := sched.At(1)

	assert.Equal(s.T(), 8, quotas["first"], "first weight is wrong")
	assert.Equal(s.T(), 2, quotas["second"], "second weight is wrong")
}

func (s *SchedulerTestSuite) TestThreeInjectorsWithDifferentWeight() {
	injectorReferences := map[string]injector.RemoteReference{
		"first": {
			Weight:    8,
			Messenger: s.bossMessenger,
		},
		"second": {
			Weight:    2,
			Messenger: s.bossMessenger,
		},
		"third": {
			Weight:    2,
			Messenger: s.bossMessenger,
		},
	}

	sched := scheduler.NewFixedIntervalScheduler(s.profile, &injectorReferences, 5)
	quotas := sched.At(1)

	assert.Equal(s.T(), 6, quotas["first"], "first weight is wrong")
	assert.Equal(s.T(), 2, quotas["second"], "second weight is wrong")
	assert.Equal(s.T(), 2, quotas["third"], "third weight is wrong")
}

func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
