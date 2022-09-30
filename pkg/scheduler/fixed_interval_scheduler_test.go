package scheduler_test

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/cockpit"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/steromano87/harkonnen/v1/pkg/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
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
		Sustain:      2 * time.Second,
		RampDown:     0,
	}

	s.bossMessenger, s.minionMessenger = messaging.NewChannelMessengerPair(100)
}

func (s *SchedulerTestSuite) TestSingleInjectorQuota() {
	injectorReferences := map[string]*injector.Reference{
		"first": {
			Weight:    1,
			Messenger: s.bossMessenger,
		},
	}

	sched := scheduler.NewFixedIntervalScheduler(s.profile, injectorReferences, 5*time.Second)
	quotas := sched.At(1 * time.Second)

	assert.Equal(s.T(), 10, quotas["first"])
}

func (s *SchedulerTestSuite) TestTwoInjectorsWithSameWeight() {
	injectorReferences := map[string]*injector.Reference{
		"first": {
			Weight:    1,
			Messenger: s.bossMessenger,
		},
		"second": {
			Weight:    1,
			Messenger: s.bossMessenger,
		},
	}
	sched := scheduler.NewFixedIntervalScheduler(s.profile, injectorReferences, 5*time.Second)
	quotas := sched.At(1 * time.Second)

	assert.Equal(s.T(), 5, quotas["first"], "first weight is wrong")
	assert.Equal(s.T(), 5, quotas["second"], "second weight is wrong")
}

func (s *SchedulerTestSuite) TestTwoInjectorsWithDifferentWeight() {
	injectorReferences := map[string]*injector.Reference{
		"first": {
			Weight:    8,
			Messenger: s.bossMessenger,
		},
		"second": {
			Weight:    2,
			Messenger: s.bossMessenger,
		},
	}

	sched := scheduler.NewFixedIntervalScheduler(s.profile, injectorReferences, 5*time.Second)
	quotas := sched.At(1 * time.Second)

	assert.Equal(s.T(), 8, quotas["first"], "first weight is wrong")
	assert.Equal(s.T(), 2, quotas["second"], "second weight is wrong")
}

func (s *SchedulerTestSuite) TestThreeInjectorsWithDifferentWeight() {
	injectorReferences := map[string]*injector.Reference{
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

	sched := scheduler.NewFixedIntervalScheduler(s.profile, injectorReferences, 5*time.Second)
	quotas := sched.At(1 * time.Second)

	assert.Equal(s.T(), 6, quotas["first"], "first weight is wrong")
	assert.Equal(s.T(), 2, quotas["second"], "second weight is wrong")
	assert.Equal(s.T(), 2, quotas["third"], "third weight is wrong")
}

func (s *SchedulerTestSuite) TestRemoteReferenceUpdate() {
	injectorReferences := map[string]*injector.Reference{
		"first": {
			Weight:    1,
			Messenger: s.bossMessenger,
		},
	}

	sched := scheduler.NewFixedIntervalScheduler(s.profile, injectorReferences, 100*time.Millisecond)

	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	ctx, cancelFunc := cockpit.NewContext(context.TODO(), &logger, project.NewConfig())
	go sched.Start(ctx)
	time.Sleep(150 * time.Millisecond)
	cancelFunc()

	assert.Equal(s.T(), 10, injectorReferences["first"].ScheduledRunners)
}

func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
