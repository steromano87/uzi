package syntheticuser_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
	"time"
)

type SyntheticUserTestSuite struct {
	suite.Suite
	ctx        context.Context
	cancelFunc context.CancelCauseFunc
	pip        pipeline.Pipeline
}

func (s *SyntheticUserTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	tempCtx, cancelFunc := context.WithCancelCause(context.TODO())
	s.ctx = logger.WithContext(tempCtx)
	s.cancelFunc = cancelFunc

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

func (s *SyntheticUserTestSuite) TestNewSyntheticUser() {
	user := syntheticuser.New(s.pip)

	if assert.IsType(s.T(), &syntheticuser.SyntheticUser{}, user) {
		assert.Equal(s.T(), syntheticuser.Ready, user.Status())
	}
}

func (s *SyntheticUserTestSuite) TestStartAndGracefulShutdown() {
	user := syntheticuser.New(s.pip)
	wg := &sync.WaitGroup{}
	var err error

	go func() {
		err = user.Run(s.ctx, wg)
	}()

	time.Sleep(10 * time.Millisecond)
	assert.Equal(s.T(), syntheticuser.Running, user.Status())

	user.RequestGracefulShutdown()
	wg.Wait()

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), syntheticuser.Stopped, user.Status())
	}
}

func (s *SyntheticUserTestSuite) TestStartAndForcedShutdown() {
	user := syntheticuser.New(s.pip)
	wg := &sync.WaitGroup{}
	var err error

	go func() {
		err = user.Run(s.ctx, wg)
	}()

	time.Sleep(10 * time.Millisecond)
	assert.Equal(s.T(), syntheticuser.Running, user.Status())

	s.cancelFunc(pipeline.ErrForcedShutdownRequested)
	wg.Wait()

	if assert.ErrorIs(s.T(), err, pipeline.ErrForcedShutdownRequested) {
		assert.Equal(s.T(), syntheticuser.Stopped, user.Status())
	}
}

func (s *SyntheticUserTestSuite) TestStatusDuringSetupPhase() {
	tempScriptContent := `
setup {
	fixed_wait {
		amount = "10s"
	}

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

	user := syntheticuser.New(decodedPipeline)
	wg := &sync.WaitGroup{}
	go func() {
		_ = user.Run(s.ctx, wg)
	}()
	defer s.cancelFunc(nil)

	time.Sleep(100 * time.Millisecond)
	assert.Equal(s.T(), syntheticuser.Starting, user.Status())
}

func (s *SyntheticUserTestSuite) TestStatusDuringTeardownPhase() {
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
	fixed_wait {
		amount = "10s"
	}

	log {
		message = "Teardown executed"
	}
}
`
	tempScript := filet.TmpFile(s.T(), "", tempScriptContent)
	defer filet.CleanUp(s.T())
	decodedPipeline, _ := pipeline.Decode([]byte(tempScriptContent), tempScript.Name())

	user := syntheticuser.New(decodedPipeline)
	wg := &sync.WaitGroup{}
	go func() {
		_ = user.Run(s.ctx, wg)
	}()
	defer s.cancelFunc(nil)

	time.Sleep(10 * time.Millisecond)
	user.RequestGracefulShutdown()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(s.T(), syntheticuser.Stopping, user.Status())
}

func (s *SyntheticUserTestSuite) TestUniqueIdForEachSynthUser() {
	user1 := syntheticuser.New(s.pip)
	user2 := syntheticuser.New(s.pip)

	assert.NotEqual(s.T(), user1.Id(), user2.Id())
}

func TestSyntheticUserTestSuite(t *testing.T) {
	suite.Run(t, new(SyntheticUserTestSuite))
}
