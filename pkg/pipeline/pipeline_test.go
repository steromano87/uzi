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

type PipelineTestSuite struct {
	suite.Suite
	l             loading.L
	ctx           *pipeline.Context
	ctxCancelFunc context.CancelFunc
}

func (s *PipelineTestSuite) SetupTest() {
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = time.RFC3339Nano
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
	sampleWriter := &MockedSampleWriter{
		Samples: []model.Sample{},
	}

	s.l = loading.L{
		Context:      context.TODO(),
		Logger:       &logger,
		Config:       project.NewEmptyConfig(),
		Variables:    loading.NewVariables(),
		SampleWriter: sampleWriter,
	}

	ctx, cancelFunc := pipeline.NewContextFromParent(s.l)
	s.ctx = ctx
	s.ctxCancelFunc = cancelFunc
}

func (s *PipelineTestSuite) TestRunnerWithFixedIterations() {
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
	decodedPipeline.MaxIterations = 3
	decodedPipeline.Start(s.ctx)
	decodedPipeline.Wait()

	if assert.Equal(s.T(), pipeline.Completed, s.ctx.Status()) {
		assert.Equal(s.T(), int64(3), s.ctx.TotalIterations())
		assert.Equal(s.T(), int64(3), s.ctx.SuccessfulIterations())
	}
}

func (s *PipelineTestSuite) TestRunnerWithPlannedShutdown() {
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
	decodedPipeline.MaxIterations = 400
	decodedPipeline.Start(s.ctx)
	time.Sleep(2 * time.Millisecond)
	s.T().Log("Asked for planned shutdown")
	s.ctx.PlannedShutdown()
	decodedPipeline.Wait()

	if assert.Equal(s.T(), pipeline.Completed, s.ctx.Status()) {
		assert.Less(s.T(), s.ctx.TotalIterations(), int64(400))
		assert.Equal(s.T(), s.ctx.TotalIterations(), s.ctx.SuccessfulIterations())
	}
}

func (s *PipelineTestSuite) TestRunnerWithGracefulShutdown() {
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
	decodedPipeline.MaxIterations = 400
	decodedPipeline.Start(s.ctx)
	time.Sleep(2 * time.Millisecond)
	s.T().Log("Asked for graceful shutdown")
	s.ctx.GracefulShutdown()
	decodedPipeline.Wait()

	if assert.Equal(s.T(), pipeline.Stopped, s.ctx.Status()) {
		assert.Less(s.T(), s.ctx.TotalIterations(), int64(400))
		assert.Equal(s.T(), s.ctx.TotalIterations(), s.ctx.SuccessfulIterations())
	}
}

func (s *PipelineTestSuite) TestRunnerWithForcedShutdown() {
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
	decodedPipeline.MaxIterations = 400
	decodedPipeline.Start(s.ctx)
	time.Sleep(2 * time.Millisecond)
	s.T().Log("Asked for forced shutdown")
	s.ctx.Terminate()
	decodedPipeline.Wait()

	if assert.Equal(s.T(), pipeline.ForcefullyStopped, s.ctx.Status()) {
		assert.Less(s.T(), s.ctx.TotalIterations(), int64(400))
		assert.Equal(s.T(), s.ctx.TotalIterations(), s.ctx.SuccessfulIterations())
	}
}

func (s *PipelineTestSuite) TestRunnerWithContextCancellation() {
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
	decodedPipeline.MaxIterations = 400
	decodedPipeline.Start(s.ctx)
	time.Sleep(2 * time.Millisecond)
	s.T().Log("Asked for context cancellation")
	s.ctxCancelFunc()
	decodedPipeline.Wait()

	if assert.Equal(s.T(), pipeline.ForcefullyStopped, s.ctx.Status()) {
		assert.Less(s.T(), s.ctx.TotalIterations(), int64(400))
		assert.Equal(s.T(), s.ctx.TotalIterations(), s.ctx.SuccessfulIterations())
	}
}

func TestPipelineTestSuite(t *testing.T) {
	suite.Run(t, new(PipelineTestSuite))
}
