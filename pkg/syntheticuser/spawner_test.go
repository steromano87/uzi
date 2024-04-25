package syntheticuser_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type SpawnerTestSuite struct {
	suite.Suite
	ctx        context.Context
	cancelFunc context.CancelCauseFunc
	pip        pipeline.Pipeline
	logger     *zerolog.Logger
}

func (s *SpawnerTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
	s.logger = &logger

	s.ctx, s.cancelFunc = context.WithCancelCause(context.TODO())

	tempScriptContent := `
setup {
	log {
		message = "Setup executed"
	}
}

main {
	fixed_wait {
		amount = "1ms"
	}

	log {
		message = "Main loop executed"
	}
}

teardown {
	log {
		message = "Teardown executed"
	}
}
`
	tempScript := filet.TmpFile(s.T(), "", tempScriptContent)
	defer filet.CleanUp(s.T())
	decodedPipeline, _ := pipeline.Decode([]byte(tempScriptContent), tempScript.Name())
	s.pip = decodedPipeline
}

func (s *SpawnerTestSuite) TearDownTest() {
	s.cancelFunc(nil)
}

func (s *SpawnerTestSuite) TestSpawnerStartedWithZeroRunningUsers() {
	spawner := syntheticuser.NewSpawner(s.pip)
	_ = spawner.SetMaxSynthUserQuota(5)
	spawner.SetLogger(s.logger)
	spawner.Serve(s.ctx)

	if assert.Zero(s.T(), spawner.ActiveUsers()) {
		assert.EqualValues(s.T(), 5, spawner.Counters().Ready)
		assert.Zero(s.T(), spawner.Counters().GracefullyShuttingDown)
		assert.Zero(s.T(), spawner.Counters().TeardownInProgress)
		assert.Zero(s.T(), spawner.Counters().Stopped)
		assert.Zero(s.T(), spawner.Counters().Error)
	}

	s.cancelFunc(nil)
	err := spawner.Wait()
	assert.NoError(s.T(), err)
}

func (s *SpawnerTestSuite) TestScaleUpToOneUser() {
	spawner := syntheticuser.NewSpawner(s.pip)
	_ = spawner.SetMaxSynthUserQuota(5)
	spawner.SetLogger(s.logger)
	spawner.Serve(s.ctx)

	err := spawner.ReconcileActiveUsers(1)
	time.Sleep(100 * time.Millisecond)
	if assert.NoError(s.T(), err) {
		assert.EqualValues(s.T(), 1, spawner.ActiveUsers())
		assert.EqualValues(s.T(), 4, spawner.Counters().Ready)
	}

	s.cancelFunc(nil)
	err = spawner.Wait()
	assert.ErrorIs(s.T(), err, context.Canceled)
}

func (s *SpawnerTestSuite) TestScaleDownFromOneUser() {
	spawner := syntheticuser.NewSpawner(s.pip)
	_ = spawner.SetMaxSynthUserQuota(5)
	spawner.SetLogger(s.logger)
	spawner.Serve(s.ctx)

	err := spawner.ReconcileActiveUsers(1)
	time.Sleep(100 * time.Millisecond)
	if assert.NoError(s.T(), err) {
		assert.EqualValues(s.T(), 1, spawner.ActiveUsers())
		assert.EqualValues(s.T(), 4, spawner.Counters().Ready)
	}

	err = spawner.ReconcileActiveUsers(0)

	if assert.NoError(s.T(), err) {
		assert.NoError(s.T(), spawner.Wait())
		assert.EqualValues(s.T(), 0, spawner.ActiveUsers())
		assert.EqualValues(s.T(), 4, spawner.Counters().Ready)
		assert.EqualValues(s.T(), 1, spawner.Counters().Stopped)
	}

	s.cancelFunc(nil)
	err = spawner.Wait()
	assert.NoError(s.T(), err)
}

func (s *SpawnerTestSuite) TestScaleUpAndDownUpToTwoUsers() {
	spawner := syntheticuser.NewSpawner(s.pip)
	_ = spawner.SetMaxSynthUserQuota(5)
	spawner.SetLogger(s.logger)
	spawner.Serve(s.ctx)

	err := spawner.ReconcileActiveUsers(2)
	time.Sleep(100 * time.Millisecond)
	if assert.NoError(s.T(), err) {
		assert.EqualValues(s.T(), 2, spawner.ActiveUsers())
		assert.EqualValues(s.T(), 3, spawner.Counters().Ready)
	}

	err = spawner.ReconcileActiveUsers(1)
	time.Sleep(100 * time.Millisecond)
	if assert.NoError(s.T(), err) {
		assert.EqualValues(s.T(), 1, spawner.ActiveUsers())
		assert.EqualValues(s.T(), 3, spawner.Counters().Ready)
		assert.EqualValues(s.T(), 1, spawner.Counters().Stopped)
	}

	s.cancelFunc(nil)
	err = spawner.Wait()
	assert.ErrorIs(s.T(), err, context.Canceled)
}

func (s *SpawnerTestSuite) TestScaleUpBeyondMaxQuota() {
	spawner := syntheticuser.NewSpawner(s.pip)
	_ = spawner.SetMaxSynthUserQuota(5)
	spawner.Serve(s.ctx)
	defer s.cancelFunc(nil)

	err := spawner.ReconcileActiveUsers(6)

	if assert.Error(s.T(), err) {
		assert.Zero(s.T(), spawner.ActiveUsers())
	}
}

func TestSpawnerTestSuite(t *testing.T) {
	suite.Run(t, new(SpawnerTestSuite))
}
