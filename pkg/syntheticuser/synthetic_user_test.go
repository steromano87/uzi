package syntheticuser_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/syntheticuser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type SyntheticUserTestSuite struct {
	suite.Suite
	ctx        dsl.Context
	cancelFunc context.CancelCauseFunc
	pip        *pipeline.Pipeline
}

func (s *SyntheticUserTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	s.ctx, s.cancelFunc = dsl.NewContext(context.TODO())
	s.ctx.Logger = &logger

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
		assert.Equal(s.T(), syntheticuser.Status_READY, user.Status())
	}
}

func (s *SyntheticUserTestSuite) TestStartAndGracefulShutdown() {
	user := syntheticuser.New(s.pip)
	errChan := make(chan error)

	go func() {
		errChan <- user.Run(s.ctx)
	}()

	time.Sleep(10 * time.Millisecond)
	assert.Equal(s.T(), syntheticuser.Status_RUNNING, user.Status())

	user.RequestGracefulShutdown()
	err := <-errChan

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), syntheticuser.Status_STOPPED, user.Status())
	}
}

func (s *SyntheticUserTestSuite) TestStartAndForcedShutdown() {
	user := syntheticuser.New(s.pip)
	errChan := make(chan error)

	go func() {
		errChan <- user.Run(s.ctx)
	}()

	time.Sleep(10 * time.Millisecond)
	assert.Equal(s.T(), syntheticuser.Status_RUNNING, user.Status())

	s.cancelFunc(pipeline.ErrForcedShutdownRequested)
	err := <-errChan

	if assert.ErrorIs(s.T(), err, pipeline.ErrForcedShutdownRequested) {
		assert.Equal(s.T(), syntheticuser.Status_STOPPED, user.Status())
	}
}

func (s *SyntheticUserTestSuite) TestStatusDuringSetupPhase() {
	tempScriptContent := `
setup {
	fixed_wait {
		amount = "500ms"
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
	errChan := make(chan error)

	go func() {
		errChan <- user.Run(s.ctx)
	}()
	defer s.cancelFunc(nil)

	time.Sleep(100 * time.Millisecond)
	status := user.Status()
	user.RequestGracefulShutdown()
	<-errChan
	assert.Equal(s.T(), syntheticuser.Status_SETUP_IN_PROGRESS, status)
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
		amount = "500ms"
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
	errChan := make(chan error)

	go func() {
		errChan <- user.Run(s.ctx)
	}()
	defer s.cancelFunc(nil)

	time.Sleep(10 * time.Millisecond)
	user.RequestGracefulShutdown()
	time.Sleep(100 * time.Millisecond)
	status := user.Status()
	<-errChan
	assert.Equal(s.T(), syntheticuser.Status_TEARDOWN_IN_PROGRESS, status)
}

func (s *SyntheticUserTestSuite) TestUniqueIdForEachSynthUser() {
	user1 := syntheticuser.New(s.pip)
	user2 := syntheticuser.New(s.pip)

	assert.NotEqual(s.T(), user1.Id(), user2.Id())
}

func (s *SyntheticUserTestSuite) TestString() {
	user := syntheticuser.New(s.pip)
	assert.Regexp(s.T(), `SyntheticUser\[id=.+, status=READY\]`, user.String())
}

func TestSyntheticUserTestSuite(t *testing.T) {
	suite.Run(t, new(SyntheticUserTestSuite))
}
