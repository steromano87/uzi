package pipeline_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
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
	ctx        messaging.Context
	cancelFunc context.CancelFunc
}

func (s *SchedulerTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	s.messenger = messaging.NewChannelMessenger(make(chan messaging.Message), make(chan messaging.Message))
	s.ctx, s.cancelFunc = messaging.NewContext(context.TODO(), &logger, s.messenger)
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

	scheduler := pipeline.NewScheduler(s.ctx)
	err := scheduler.Prepare(decodedPipeline, 10, int64(999))

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), 10, scheduler.Stats().Ready)
		assert.Equal(s.T(), 0, scheduler.Stats().Running)
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

	scheduler := pipeline.NewScheduler(s.ctx)
	_ = scheduler.Prepare(decodedPipeline, 3, int64(999))
	err := scheduler.Schedule(1)

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), 1, scheduler.Stats().Running)
	}

	time.Sleep(2 * time.Millisecond)
	err = scheduler.Schedule(0)

	if assert.NoError(s.T(), err) {
		scheduler.WaitForCompletion()
		assert.Equal(s.T(), 0, scheduler.Stats().Running)
		assert.Equal(s.T(), 1, scheduler.Stats().Completed)
	}
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

	scheduler := pipeline.NewScheduler(s.ctx)
	_ = scheduler.Prepare(decodedPipeline, 3, int64(5))
	err := scheduler.Schedule(3)

	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), 3, scheduler.Stats().Running)
	}

	scheduler.WaitForCompletion()
	if assert.Equal(s.T(), 0, scheduler.Stats().Running) {
		assert.Equal(s.T(), 3, scheduler.Stats().Completed)
	}
}

func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(SchedulerTestSuite))
}
