package loading_test

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
	"time"
)

type ShooterTestSuite struct {
	suite.Suite
	l            loading.L
	sampleWriter *MockedSampleWriter
	mainScript   loading.Script
	wg           sync.WaitGroup
}

func (s *ShooterTestSuite) SetupTest() {
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = time.RFC3339Nano
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
	s.sampleWriter = &MockedSampleWriter{
		Samples: []model.Sample{},
	}
	s.l = loading.L{
		Context:      context.TODO(),
		Logger:       &logger,
		Config:       project.NewEmptyConfig(),
		Variables:    loading.NewVariables(),
		SampleWriter: s.sampleWriter,
	}

	s.mainScript = func(l loading.L) error {
		l.Logger.Info().Msg("Main loop")
		return nil
	}
	s.wg = sync.WaitGroup{}
}

func (s *ShooterTestSuite) TestShooterSyncRunWithOnlyOneMainScript() {
	shooter := loading.Shooter{
		ID:             "one",
		SetUpScript:    nil,
		MainScripts:    []loading.Script{s.mainScript},
		TearDownScript: nil,
		MaxIterations:  3,
	}

	shooter.Run(s.l)
	if assert.Equal(s.T(), loading.ShooterStatusCompleted, shooter.Status()) {
		assert.EqualValues(s.T(), 3, shooter.TotalIterations())
	}
}

func (s *ShooterTestSuite) TestShooterAsyncRunWithOnlyOneMainScript() {
	shooter := loading.Shooter{
		ID:             "one",
		SetUpScript:    nil,
		MainScripts:    []loading.Script{s.mainScript},
		TearDownScript: nil,
		MaxIterations:  3,
	}

	s.wg.Add(1)
	shooter.RunAsync(s.l, &s.wg)
	s.wg.Wait()
	if assert.Equal(s.T(), loading.ShooterStatusCompleted, shooter.Status()) {
		assert.EqualValues(s.T(), 3, shooter.TotalIterations())
	}
}

func (s *ShooterTestSuite) TestShooterAsyncRunWithGracefulShutdown() {
	shooter := loading.Shooter{
		ID:             "one",
		SetUpScript:    nil,
		MainScripts:    []loading.Script{s.mainScript},
		TearDownScript: nil,
		MaxIterations:  0,
	}

	s.wg.Add(1)
	shooter.RunAsync(s.l, &s.wg)
	time.Sleep(2 * time.Millisecond)
	shooter.GracefulShutdown()
	s.wg.Wait()
	if assert.Equal(s.T(), loading.ShooterStatusStopped, shooter.Status()) {
		assert.Greater(s.T(), shooter.TotalIterations(), int64(0))
		assert.Less(s.T(), shooter.TotalIterations(), int64(999))
	}
}

func (s *ShooterTestSuite) TestShooterAsyncRunWithForcedShutdown() {
	shooter := loading.Shooter{
		ID:             "one",
		SetUpScript:    nil,
		MainScripts:    []loading.Script{s.mainScript},
		TearDownScript: nil,
		MaxIterations:  999,
	}

	s.wg.Add(1)
	shooter.RunAsync(s.l, &s.wg)
	time.Sleep(2 * time.Millisecond)
	shooter.ForcedShutdown()
	s.wg.Wait()
	if assert.Equal(s.T(), loading.ShooterStatusForcefullyStopped, shooter.Status()) {
		assert.Greater(s.T(), shooter.TotalIterations(), int64(0))
		assert.Less(s.T(), shooter.TotalIterations(), int64(999))
	}
}

func TestShooterTestSuite(t *testing.T) {
	suite.Run(t, new(ShooterTestSuite))
}
