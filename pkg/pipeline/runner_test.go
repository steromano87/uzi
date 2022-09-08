package pipeline_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
	"time"
)

type RunnerTestSuite struct {
	suite.Suite
	backgroundCtx        context.Context
	backGroundCancelFunc context.CancelFunc
	ctx                  *pipeline.Context
	cancelFunc           context.CancelFunc
	messenger            messaging.Messenger
	iterCounter          *pipeline.IterationsCounter
}

func (s *RunnerTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	s.messenger = messaging.NewChannelMessenger(make(chan messaging.Message), make(chan messaging.Message))
	s.iterCounter = pipeline.NewIterationsCounter()

	s.backgroundCtx = context.TODO()
	newCtx, cancelFunc := context.WithCancel(s.backgroundCtx)
	s.backGroundCancelFunc = cancelFunc
	s.ctx, s.cancelFunc = pipeline.NewContext(newCtx, project.NewConfig(), &logger, s.messenger, s.iterCounter)
}

func (s *RunnerTestSuite) TestRunnerWithFixedIterations() {
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

	waitGroup := sync.WaitGroup{}

	decodedPipeline, _ := pipeline.Decode([]byte(tempScriptContent), tempScript.Name())
	s.ctx.IterationsCounter().SetMaxIterations(3)
	runner := pipeline.NewRunner(s.ctx)
	waitGroup.Add(1)
	runner.Start(&waitGroup, decodedPipeline)
	waitGroup.Wait()

	if assert.Equal(s.T(), pipeline.Completed, s.ctx.Status()) {
		assert.EqualValues(s.T(), 3, s.ctx.IterationsCounter().CompletedIterations())
		assert.EqualValues(s.T(), 3, s.ctx.IterationsCounter().PassedIterations())
	}
}

func (s *RunnerTestSuite) TestRunnerWithPlannedShutdown() {
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

	waitGroup := sync.WaitGroup{}

	decodedPipeline, _ := pipeline.Decode([]byte(tempScriptContent), tempScript.Name())
	runner := pipeline.NewRunner(s.ctx)
	s.ctx.IterationsCounter().SetMaxIterations(9999)
	waitGroup.Add(1)
	runner.Start(&waitGroup, decodedPipeline)
	time.Sleep(2 * time.Millisecond)
	s.T().Log("Asked for planned shutdown")
	s.ctx.SchedulePlannedShutdown()
	waitGroup.Wait()

	if assert.Equal(s.T(), pipeline.Completed, s.ctx.Status()) {
		assert.Less(s.T(), s.ctx.IterationsCounter().CompletedIterations(), uint64(9999))
		assert.Equal(s.T(), s.ctx.IterationsCounter().CompletedIterations(), s.ctx.IterationsCounter().PassedIterations())
	}
}

func (s *RunnerTestSuite) TestRunnerWithGracefulShutdown() {
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

	waitGroup := sync.WaitGroup{}

	decodedPipeline, _ := pipeline.Decode([]byte(tempScriptContent), tempScript.Name())
	runner := pipeline.NewRunner(s.ctx)
	s.ctx.IterationsCounter().SetMaxIterations(9999)
	waitGroup.Add(1)
	runner.Start(&waitGroup, decodedPipeline)
	time.Sleep(2 * time.Millisecond)
	s.T().Log("Asked for planned shutdown")
	s.ctx.ScheduleGracefulShutdown()
	waitGroup.Wait()

	if assert.Equal(s.T(), pipeline.Stopped, s.ctx.Status()) {
		assert.Less(s.T(), s.ctx.IterationsCounter().CompletedIterations(), uint64(9999))
		assert.Equal(s.T(), s.ctx.IterationsCounter().CompletedIterations(), s.ctx.IterationsCounter().PassedIterations())
	}
}

func (s *RunnerTestSuite) TestRunnerWithForcedShutdown() {
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

	waitGroup := sync.WaitGroup{}

	decodedPipeline, _ := pipeline.Decode([]byte(tempScriptContent), tempScript.Name())
	runner := pipeline.NewRunner(s.ctx)
	s.ctx.IterationsCounter().SetMaxIterations(9999)
	waitGroup.Add(1)
	runner.Start(&waitGroup, decodedPipeline)
	time.Sleep(2 * time.Millisecond)
	s.T().Log("Asked for forced shutdown")
	s.cancelFunc()
	waitGroup.Wait()

	if assert.Equal(s.T(), pipeline.ForcefullyStopped, s.ctx.Status()) {
		assert.Less(s.T(), s.ctx.IterationsCounter().CompletedIterations(), uint64(9999))
		assert.Equal(s.T(), s.ctx.IterationsCounter().CompletedIterations(), s.ctx.IterationsCounter().PassedIterations())
	}
}

func (s *RunnerTestSuite) TestRunnerWithParentContextCancellation() {
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

	waitGroup := sync.WaitGroup{}

	decodedPipeline, _ := pipeline.Decode([]byte(tempScriptContent), tempScript.Name())
	runner := pipeline.NewRunner(s.ctx)
	s.ctx.IterationsCounter().SetMaxIterations(9999)
	waitGroup.Add(1)
	runner.Start(&waitGroup, decodedPipeline)
	time.Sleep(2 * time.Millisecond)
	s.T().Log("Asked for context cancellation")
	s.backGroundCancelFunc()
	waitGroup.Wait()

	if assert.Equal(s.T(), pipeline.ForcefullyStopped, s.ctx.Status()) {
		assert.Less(s.T(), s.ctx.IterationsCounter().CompletedIterations(), uint64(9999))
		assert.Equal(s.T(), s.ctx.IterationsCounter().CompletedIterations(), s.ctx.IterationsCounter().PassedIterations())
	}
}

func TestPipelineTestSuite(t *testing.T) {
	suite.Run(t, new(RunnerTestSuite))
}
