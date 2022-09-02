package injector_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type SchedulerTestSuite struct {
	suite.Suite
	messenger  messaging.Messenger
	ctx        injector.Context
	cancelFunc context.CancelFunc
}

func (s *SchedulerTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	s.messenger = messaging.NewChannelMessenger(make(chan messaging.Message, 9999), make(chan messaging.Message, 9999))
	s.ctx, s.cancelFunc = injector.NewContext(context.TODO(), &logger, s.messenger)
}

func (s *SchedulerTestSuite) TestSchedulePreparation() {
	tempScriptContent := `
setup {
	log {
		message = "Setup executed"
	}
}

main {
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

	dispatcher := injector.NewRunnerDispatcher(s.ctx)
	err := dispatcher.Prepare(decodedPipeline, 10, 999)

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), 10, dispatcher.Stats().Ready)
		assert.Equal(s.T(), 0, dispatcher.Stats().Running)
	}
}

func (s *SchedulerTestSuite) TestPipelineStartAndPlannedShutdown() {
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

	dispatcher := injector.NewRunnerDispatcher(s.ctx)
	_ = dispatcher.Prepare(decodedPipeline, 3, 999)

	err := dispatcher.Dispatch(1)

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), 1, dispatcher.Stats().Running)
	}

	time.Sleep(2 * time.Millisecond)
	err = dispatcher.Dispatch(0)

	if assert.NoError(s.T(), err) {
		dispatcher.WaitForCompletion()
		assert.Equal(s.T(), 0, dispatcher.Stats().Running)
		assert.Equal(s.T(), 1, dispatcher.Stats().Completed)
	}
}

func (s *SchedulerTestSuite) TestGracefulShutdown() {
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

	dispatcher := injector.NewRunnerDispatcher(s.ctx)
	_ = dispatcher.Prepare(decodedPipeline, 3, 999)

	err := dispatcher.Dispatch(3)

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), 3, dispatcher.Stats().Running)
	}

	time.Sleep(2 * time.Millisecond)
	dispatcher.GracefulShutdown()
	dispatcher.WaitForCompletion()
	assert.Equal(s.T(), 0, dispatcher.Stats().Running)
	assert.Equal(s.T(), 3, dispatcher.Stats().Stopped)
}

func (s *SchedulerTestSuite) TestPipelineStartWithFixedIterations() {
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

	dispatcher := injector.NewRunnerDispatcher(s.ctx)
	_ = dispatcher.Prepare(decodedPipeline, 3, 5)
	err := dispatcher.Dispatch(3)

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), 3, dispatcher.Stats().Running)
	}

	dispatcher.WaitForCompletion()
	if assert.Equal(s.T(), 0, dispatcher.Stats().Running) {
		assert.Equal(s.T(), 3, dispatcher.Stats().Completed)
		assert.EqualValues(s.T(), 5, dispatcher.IterationsCounter().CompletedIterations())
		assert.EqualValues(s.T(), 5, dispatcher.IterationsCounter().PassedIterations())
	}
}

func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
