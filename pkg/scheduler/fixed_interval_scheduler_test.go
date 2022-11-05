package scheduler_test

import (
	"context"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type SchedulerTestSuite struct {
	suite.Suite
	profile     scheduler.LoadProfile
	grpcChannel *inprocgrpc.Channel
	logger      *zerolog.Logger
}

func (s *SchedulerTestSuite) SetupTest() {
	s.profile = scheduler.LinearRamp{
		Runners:      10,
		InitialDelay: 0,
		RampUp:       0,
		Sustain:      2 * time.Second,
		RampDown:     0,
	}
	s.grpcChannel = &inprocgrpc.Channel{}
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
	s.logger = &logger
}

func (s *SchedulerTestSuite) TestSingleInjectorQuota() {
	injectorReferences := map[string]*injector.Reference{
		"first": {
			Weight:         1,
			InjectorClient: injector.NewInjectorClient(s.grpcChannel),
		},
	}

	sched := scheduler.NewFixedIntervalScheduler(s.logger, s.profile, injectorReferences, 5*time.Second)
	quotas := sched.At(1 * time.Second)

	assert.EqualValues(s.T(), 10, quotas["first"])
}

func (s *SchedulerTestSuite) TestTwoInjectorsWithSameWeight() {
	injectorReferences := map[string]*injector.Reference{
		"first": {
			Weight:         1,
			InjectorClient: injector.NewInjectorClient(s.grpcChannel),
		},
		"second": {
			Weight:         1,
			InjectorClient: injector.NewInjectorClient(s.grpcChannel),
		},
	}
	sched := scheduler.NewFixedIntervalScheduler(s.logger, s.profile, injectorReferences, 5*time.Second)
	quotas := sched.At(1 * time.Second)

	assert.EqualValues(s.T(), 5, quotas["first"], "first weight is wrong")
	assert.EqualValues(s.T(), 5, quotas["second"], "second weight is wrong")
}

func (s *SchedulerTestSuite) TestTwoInjectorsWithDifferentWeight() {
	injectorReferences := map[string]*injector.Reference{
		"first": {
			Weight:         8,
			InjectorClient: injector.NewInjectorClient(s.grpcChannel),
		},
		"second": {
			Weight:         2,
			InjectorClient: injector.NewInjectorClient(s.grpcChannel),
		},
	}

	sched := scheduler.NewFixedIntervalScheduler(s.logger, s.profile, injectorReferences, 5*time.Second)
	quotas := sched.At(1 * time.Second)

	assert.EqualValues(s.T(), 8, quotas["first"], "first weight is wrong")
	assert.EqualValues(s.T(), 2, quotas["second"], "second weight is wrong")
}

func (s *SchedulerTestSuite) TestThreeInjectorsWithDifferentWeight() {
	injectorReferences := map[string]*injector.Reference{
		"first": {
			Weight:         8,
			InjectorClient: injector.NewInjectorClient(s.grpcChannel),
		},
		"second": {
			Weight:         2,
			InjectorClient: injector.NewInjectorClient(s.grpcChannel),
		},
		"third": {
			Weight:         2,
			InjectorClient: injector.NewInjectorClient(s.grpcChannel),
		},
	}

	sched := scheduler.NewFixedIntervalScheduler(s.logger, s.profile, injectorReferences, 5*time.Second)
	quotas := sched.At(1 * time.Second)

	assert.EqualValues(s.T(), 6, quotas["first"], "first weight is wrong")
	assert.EqualValues(s.T(), 2, quotas["second"], "second weight is wrong")
	assert.EqualValues(s.T(), 2, quotas["third"], "third weight is wrong")
}

func (s *SchedulerTestSuite) TestRemoteReferenceUpdate() {
	injectorReferences := map[string]*injector.Reference{
		"first": {
			Weight:         1,
			InjectorClient: injector.NewInjectorClient(s.grpcChannel),
		},
	}

	sched := scheduler.NewFixedIntervalScheduler(s.logger, s.profile, injectorReferences, 100*time.Millisecond)

	ctx, cancelFunc := context.WithCancel(context.TODO())
	go sched.Start(ctx)
	time.Sleep(150 * time.Millisecond)
	cancelFunc()

	assert.EqualValues(s.T(), 10, injectorReferences["first"].ScheduledRunners)
}

func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
