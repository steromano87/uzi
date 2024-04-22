package pipeline_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

/////////////////////////////////////
// Mocked telemetry sinks for test //
/////////////////////////////////////

type TestLogSink struct {
	Logs []string
}

func NewTestLogSink() *TestLogSink {
	tls := new(TestLogSink)
	tls.Logs = make([]string, 0)

	return tls
}

func (tls *TestLogSink) Write(p []byte) (n int, err error) {
	tls.Logs = append(tls.Logs, string(p))
	return len(p), nil
}

////////////////
// Test suite //
////////////////

type PipelineTestSuite struct {
	suite.Suite
	noOpPipeline pipeline.Pipeline
	logSink      *TestLogSink
	logger       zerolog.Logger
}

func (s *PipelineTestSuite) SetupTest() {
	tempScriptContent := `
setup {
	log {
		message = "Setup done"
	}
}

main {
	log {
		message = "Main step 1 done"
	}

	log {
		message = "Main step 2 done"
	}
}

teardown {
	log {
		message = "Teardown done"
	}
}
`
	tempScript := filet.TmpFile(s.T(), "", tempScriptContent)
	defer filet.CleanUp(s.T())

	s.noOpPipeline, _ = pipeline.Decode([]byte(tempScriptContent), tempScript.Name())
	s.logSink = NewTestLogSink()

	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02T15:04:05.000"}
	multiWriter := zerolog.MultiLevelWriter(consoleWriter, s.logSink)

	s.logger = zerolog.New(multiWriter).With().Timestamp().Logger()
}

func (s *PipelineTestSuite) TestPipeline_RunSetup() {
	ctx, _ := dsl.NewContext(context.TODO())
	ctx.Logger = &s.logger
	err := s.noOpPipeline.RunSetup(ctx)

	if assert.NoError(s.T(), err) {
		// Expected length is 3 because of start and end logs
		if assert.Len(s.T(), s.logSink.Logs, 3) {
			logEntry := s.logSink.Logs[1]
			assert.Contains(s.T(), logEntry, "Setup done")
		}
	}
}

func (s *PipelineTestSuite) TestPipeline_RunTeardown() {
	ctx, _ := dsl.NewContext(context.TODO())
	ctx.Logger = &s.logger
	err := s.noOpPipeline.RunTeardown(ctx)

	if assert.NoError(s.T(), err) {
		// Expected length is 3 because of start and end logs
		if assert.Len(s.T(), s.logSink.Logs, 3) {
			logEntry := s.logSink.Logs[1]
			assert.Contains(s.T(), logEntry, "Teardown done")
		}
	}
}

func (s *PipelineTestSuite) TestPipeline_RunMainOnce() {
	ctx, _ := dsl.NewContext(context.TODO())
	ctx.Logger = &s.logger
	err := s.noOpPipeline.RunMainOnce(ctx)

	if assert.NoError(s.T(), err) {
		// Expected length is 6 because of start and end logs for every step
		if assert.Len(s.T(), s.logSink.Logs, 6) {
			assert.Contains(s.T(), s.logSink.Logs[1], "Main step 1 done")
			assert.Contains(s.T(), s.logSink.Logs[4], "Main step 2 done")
		}
	}
}

func (s *PipelineTestSuite) TestPipeline_RunMainWithGracefulShutdown() {
	ctx, _ := dsl.NewContext(context.TODO())
	ctx.Logger = &s.logger
	errorChan := make(chan error)

	// Run pipeline in separate goroutine
	go func() {
		errorChan <- s.noOpPipeline.RunMain(ctx)
	}()

	time.Sleep(10 * time.Millisecond)
	s.noOpPipeline.RequestGracefulShutdown()
	err := <-errorChan

	if assert.NoError(s.T(), err) {
		// Expected length is 3 because of start and end logs
		if assert.Greater(s.T(), len(s.logSink.Logs), 6) {
			logEntry := s.logSink.Logs[len(s.logSink.Logs)-1]
			assert.Contains(s.T(), logEntry, "Gracefully exited main loop")
		}
	}
}

func (s *PipelineTestSuite) TestPipeline_RunMainWithForcedShutdown() {
	ctx, cancelFunc := dsl.NewContext(context.TODO())
	ctx.Logger = &s.logger
	errorChan := make(chan error)

	// Run pipeline in separate goroutine
	go func() {
		errorChan <- s.noOpPipeline.RunMain(ctx)
	}()

	time.Sleep(10 * time.Millisecond)
	cancelFunc(nil)
	err := <-errorChan

	if assert.ErrorIs(s.T(), err, context.Canceled) {
		// Expected length is 3 because of start and end logs
		if assert.Greater(s.T(), len(s.logSink.Logs), 6) {
			logEntry := s.logSink.Logs[len(s.logSink.Logs)-1]
			assert.Contains(s.T(), logEntry, "Step run interrupted")
		}
	}
}

func (s *PipelineTestSuite) TestPipeline_RunMainWithGracefulShutdownAndError() {
	failingScriptContent := `
main {
	fail {
		message = "Thou shall not pass"
	}
}
`
	tempScript := filet.TmpFile(s.T(), "", failingScriptContent)
	defer filet.CleanUp(s.T())

	failingPipeline, _ := pipeline.Decode([]byte(failingScriptContent), tempScript.Name())

	ctx, _ := dsl.NewContext(context.TODO())
	ctx.Logger = &s.logger
	errorChan := make(chan error)

	// Run pipeline in separate goroutine
	go func() {
		errorChan <- failingPipeline.RunMain(ctx)
	}()

	time.Sleep(10 * time.Millisecond)
	failingPipeline.RequestGracefulShutdown()
	err := <-errorChan

	if assert.NoError(s.T(), err) {
		assert.Zero(s.T(), failingPipeline.PassedIterations())
		assert.EqualValues(s.T(), failingPipeline.CompletedIterations(), failingPipeline.FailedIterations())
	}
}

func (s *PipelineTestSuite) TestPipeline_RunMainWithMaxIterations() {
	ctx, _ := dsl.NewContext(context.TODO())
	ctx.Logger = &s.logger
	s.noOpPipeline.MaxIterations = 5
	errorChan := make(chan error)

	// Run pipeline in separate goroutine
	go func() {
		errorChan <- s.noOpPipeline.RunMain(ctx)
	}()

	err := <-errorChan

	if assert.NoError(s.T(), err) {
		assert.EqualValues(s.T(), 5, s.noOpPipeline.CompletedIterations())
		assert.EqualValues(s.T(), 5, s.noOpPipeline.PassedIterations())
		assert.Zero(s.T(), s.noOpPipeline.InProgressIterations())
		assert.Zero(s.T(), s.noOpPipeline.FailedIterations())
	}
}

func TestPipelineTestSuite(t *testing.T) {
	suite.Run(t, new(PipelineTestSuite))
}
