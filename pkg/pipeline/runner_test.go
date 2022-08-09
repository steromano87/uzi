package pipeline_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/loading"
	"github.com/steromano87/harkonnen/v1/pkg/model"
	"github.com/steromano87/harkonnen/v1/pkg/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
	"time"
)

type MockedSampleWriter struct {
	Samples []model.Sample
}

func (w *MockedSampleWriter) Write(sample model.Sample) error {
	w.Samples = append(w.Samples, sample)
	return nil
}

type RunnerTestSuite struct {
	suite.Suite
	l             loading.L
	backgroundCtx context.Context
	cancelFunc    context.CancelFunc
}

func (s *RunnerTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
	sampleWriter := &MockedSampleWriter{
		Samples: []model.Sample{},
	}

	s.backgroundCtx = context.TODO()
	newCtx, cancelFunc := context.WithCancel(s.backgroundCtx)
	s.cancelFunc = cancelFunc

	s.l = loading.L{
		Context:      newCtx,
		Logger:       &logger,
		Config:       project.NewEmptyConfig(),
		Variables:    loading.NewVariables(),
		SampleWriter: sampleWriter,
	}
}

func (s *RunnerTestSuite) TestRunnerWithFixedIterations() {
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

	waitGroup := sync.WaitGroup{}

	decodedPipeline, _ := pipeline.Decode([]byte(tempScriptContent), tempScript.Name())
	runner := pipeline.NewRunner(s.l, decodedPipeline, 3)
	waitGroup.Add(1)
	runner.Start(&waitGroup)
	waitGroup.Wait()

	if assert.Equal(s.T(), pipeline.Completed, runner.Status()) {
		assert.Equal(s.T(), int64(3), runner.TotalIterations())
		assert.Equal(s.T(), int64(3), runner.SuccessfulIterations())
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
	runner := pipeline.NewRunner(s.l, decodedPipeline, 9999)
	waitGroup.Add(1)
	runner.Start(&waitGroup)
	time.Sleep(2 * time.Millisecond)
	s.T().Log("Asked for planned shutdown")
	runner.PlannedShutdown()
	waitGroup.Wait()

	if assert.Equal(s.T(), pipeline.Completed, runner.Status()) {
		assert.Less(s.T(), runner.TotalIterations(), int64(9999))
		assert.Equal(s.T(), runner.TotalIterations(), runner.SuccessfulIterations())
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
	runner := pipeline.NewRunner(s.l, decodedPipeline, 9999)
	waitGroup.Add(1)
	runner.Start(&waitGroup)
	time.Sleep(2 * time.Millisecond)
	s.T().Log("Asked for graceful shutdown")
	runner.GracefulShutdown()
	waitGroup.Wait()

	if assert.Equal(s.T(), pipeline.Stopped, runner.Status()) {
		assert.Less(s.T(), runner.TotalIterations(), int64(9999))
		assert.Equal(s.T(), runner.TotalIterations(), runner.SuccessfulIterations())
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
	runner := pipeline.NewRunner(s.l, decodedPipeline, 9999)
	waitGroup.Add(1)
	runner.Start(&waitGroup)
	time.Sleep(2 * time.Millisecond)
	s.T().Log("Asked for forced shutdown")
	runner.Terminate()
	waitGroup.Wait()

	if assert.Equal(s.T(), pipeline.ForcefullyStopped, runner.Status()) {
		assert.Less(s.T(), runner.TotalIterations(), int64(9999))
		assert.Equal(s.T(), runner.TotalIterations(), runner.SuccessfulIterations())
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
	runner := pipeline.NewRunner(s.l, decodedPipeline, 9999)
	waitGroup.Add(1)
	runner.Start(&waitGroup)
	time.Sleep(2 * time.Millisecond)
	s.T().Log("Asked for context cancellation")
	s.cancelFunc()
	waitGroup.Wait()

	if assert.Equal(s.T(), pipeline.ForcefullyStopped, runner.Status()) {
		assert.Less(s.T(), runner.TotalIterations(), int64(9999))
		assert.Equal(s.T(), runner.TotalIterations(), runner.SuccessfulIterations())
	}
}

func TestPipelineTestSuite(t *testing.T) {
	suite.Run(t, new(RunnerTestSuite))
}
