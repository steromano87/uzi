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

type JITSchedulerTestSuite struct {
	suite.Suite
	l loading.L
}

func (s *JITSchedulerTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
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
}

func (s *JITSchedulerTestSuite) TestSchedulePreparation() {
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

	scheduler := pipeline.NewScheduler(s.l)
	err := scheduler.Prepare(decodedPipeline, 10, int64(999))

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), 0, scheduler.RunningPipelines())
	}
}

func (s *JITSchedulerTestSuite) TestPipelineStartAndPlannedShutdown() {
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

	scheduler := pipeline.NewScheduler(s.l)
	_ = scheduler.Prepare(decodedPipeline, 3, int64(999))
	err := scheduler.Schedule(1)

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), 1, scheduler.RunningPipelines())
	}

	time.Sleep(2 * time.Millisecond)
	err = scheduler.Schedule(0)

	if assert.NoError(s.T(), err) {
		scheduler.WaitForCompletion()
		assert.Equal(s.T(), 0, scheduler.RunningPipelines())
	}
}

func (s *JITSchedulerTestSuite) TestPipelineStartWithFixedIterations() {
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

	scheduler := pipeline.NewScheduler(s.l)
	_ = scheduler.Prepare(decodedPipeline, 3, int64(25))
	err := scheduler.Schedule(3)

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), 3, scheduler.RunningPipelines())
	}

	scheduler.WaitForCompletion()
	assert.Equal(s.T(), 0, scheduler.RunningPipelines())
}

func TestJITSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(JITSchedulerTestSuite))
}
