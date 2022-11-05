package injector_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/configuration"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type RunnerPoolTestSuite struct {
	suite.Suite
	logger                *zerolog.Logger
	configuration         *configuration.Configuration
	ctx                   context.Context
	cancelFunc            context.CancelFunc
	injectorClient        injector.InjectorClient
	injectorInProcChannel *inprocgrpc.Channel
	telemetryServer       *telemetry.Server
}

func (s *RunnerPoolTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()
	s.logger = &logger

	s.configuration, _ = configuration.NewDefault()

	s.ctx, s.cancelFunc = context.WithCancel(context.TODO())
	s.injectorInProcChannel = &inprocgrpc.Channel{}
	s.injectorClient = injector.NewInjectorClient(s.injectorInProcChannel)
	s.telemetryServer = telemetry.NewServer(s.configuration)
}

func (s *RunnerPoolTestSuite) TestSchedulePreparation() {
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

	runnerPool := injector.RunnerPool{
		Configuration:    s.configuration,
		Logger:           s.logger,
		TemplatePipeline: decodedPipeline,
		TelemetryServer:  s.telemetryServer,
	}
	runnerPool.Initialize()
	err := runnerPool.SetDesiredRunners(s.ctx, 10)

	if assert.NoError(s.T(), err) {
		assert.EqualValues(s.T(), 10, runnerPool.DispatchedRunners())
	}
}

func (s *RunnerPoolTestSuite) TestPipelineStartAndPlannedShutdown() {
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

	runnerPool := injector.RunnerPool{
		Configuration:    s.configuration,
		Logger:           s.logger,
		TemplatePipeline: decodedPipeline,
		TelemetryServer:  s.telemetryServer,
	}
	runnerPool.Initialize()
	err := runnerPool.SetDesiredRunners(s.ctx, 1)

	if assert.NoError(s.T(), err) {
		assert.EqualValues(s.T(), 1, runnerPool.DispatchedRunners())
	}

	time.Sleep(2 * time.Millisecond)
	err = runnerPool.SetDesiredRunners(s.ctx, 0)

	if assert.NoError(s.T(), err) {
		runnerPool.WaitForCompletion()
		assert.EqualValues(s.T(), 0, runnerPool.DispatchedRunners())
	}
}

func (s *RunnerPoolTestSuite) TestGracefulShutdown() {
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

	runnerPool := injector.RunnerPool{
		Configuration:    s.configuration,
		Logger:           s.logger,
		TemplatePipeline: decodedPipeline,
		TelemetryServer:  s.telemetryServer,
	}
	runnerPool.Initialize()
	err := runnerPool.SetDesiredRunners(s.ctx, 3)

	if assert.NoError(s.T(), err) {
		assert.EqualValues(s.T(), 3, runnerPool.DispatchedRunners())
	}

	time.Sleep(2 * time.Millisecond)
	runnerPool.GracefulShutdown()
	runnerPool.WaitForCompletion()
	if assert.EqualValues(s.T(), 0, runnerPool.DispatchedRunners()) {
		assert.Equal(s.T(), s.telemetryServer.GetCounters().GetCompleted(), s.telemetryServer.GetCounters().GetPassed())
	}
}

func (s *RunnerPoolTestSuite) TestPipelineStartWithFixedIterations() {
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
	s.configuration.Load.MaxIterations = 5
	s.telemetryServer = telemetry.NewServer(s.configuration)

	runnerPool := injector.RunnerPool{
		Configuration:    s.configuration,
		Logger:           s.logger,
		TemplatePipeline: decodedPipeline,
		TelemetryServer:  s.telemetryServer,
	}
	runnerPool.Initialize()
	err := runnerPool.SetDesiredRunners(s.ctx, 3)

	if assert.NoError(s.T(), err) {
		assert.EqualValues(s.T(), 3, runnerPool.DispatchedRunners())
	}

	runnerPool.WaitForCompletion()
	if assert.EqualValues(s.T(), 0, runnerPool.DispatchedRunners()) {
		assert.EqualValues(s.T(), 5, s.telemetryServer.GetCounters().GetCompleted())
		assert.EqualValues(s.T(), 5, s.telemetryServer.GetCounters().GetPassed())
		assert.EqualValues(s.T(), 0, s.telemetryServer.GetCounters().GetFailed())
		assert.EqualValues(s.T(), 0, s.telemetryServer.GetCounters().GetInProgress())
	}
}

func TestRunnerPoolTestSuite(t *testing.T) {
	suite.Run(t, new(RunnerPoolTestSuite))
}
