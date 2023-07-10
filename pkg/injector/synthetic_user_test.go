package injector_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/dsl/pipeline"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

/////////////////////////
// Test harness begin  //
/////////////////////////

type statusCollector struct {
	oldStatuses []injector.SyntheticUserStatus
	newStatuses []injector.SyntheticUserStatus
}

func newStatusCollector() statusCollector {
	return statusCollector{
		oldStatuses: make([]injector.SyntheticUserStatus, 0),
		newStatuses: make([]injector.SyntheticUserStatus, 0),
	}
}

func (sc *statusCollector) onStatusChange(oldStatus, newStatus injector.SyntheticUserStatus) error {
	sc.oldStatuses = append(sc.oldStatuses, oldStatus)
	sc.newStatuses = append(sc.newStatuses, newStatus)
	return nil
}

/////////////////////////
// Test harness end    //
/////////////////////////

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
	user := injector.NewSyntheticUser()

	if assert.IsType(s.T(), &injector.SyntheticUser{}, user) {
		assert.Equal(s.T(), injector.SyntheticUserStatus_READY, user.Status())
	}
}

func (s *SyntheticUserTestSuite) TestStartAndGracefulShutdown() {
	user := injector.NewSyntheticUser()
	user.Run(s.ctx, s.pip)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(s.T(), injector.SyntheticUserStatus_RUNNING, user.Status())

	user.StartGracefulShutdown()
	user.Wait()

	assert.Equal(s.T(), injector.SyntheticUserStatus_STOPPED, user.Status())
}

func (s *SyntheticUserTestSuite) TestStartAndForcedShutdown() {
	user := injector.NewSyntheticUser()
	user.Run(s.ctx, s.pip)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(s.T(), injector.SyntheticUserStatus_RUNNING, user.Status())

	s.cancelFunc(injector.ErrForcedShutdownRequested)
	user.Wait()

	assert.Equal(s.T(), injector.SyntheticUserStatus_STOPPED, user.Status())
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

	user := injector.NewSyntheticUser()
	user.Run(s.ctx, decodedPipeline)
	defer s.cancelFunc(context.Canceled)
	time.Sleep(100 * time.Millisecond)
	assert.Equal(s.T(), injector.SyntheticUserStatus_STARTING, user.Status())
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

	user := injector.NewSyntheticUser()
	user.Run(s.ctx, decodedPipeline)
	defer s.cancelFunc(context.Canceled)
	time.Sleep(10 * time.Millisecond)
	user.StartGracefulShutdown()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(s.T(), injector.SyntheticUserStatus_STOPPING, user.Status())
}

func (s *SyntheticUserTestSuite) TestStatusChangeCallback() {
	statusCollector := newStatusCollector()
	user := injector.NewSyntheticUser()
	user.RegisterStatusChangeFunc(statusCollector.onStatusChange)
	user.Run(s.ctx, s.pip)
	time.Sleep(10 * time.Millisecond)

	user.StartGracefulShutdown()
	user.Wait()

	if assert.NotEmpty(s.T(), statusCollector.newStatuses) {
		assert.Len(s.T(), statusCollector.oldStatuses, 4)
		assert.Len(s.T(), statusCollector.newStatuses, 4)
	}
}

func (s *SyntheticUserTestSuite) TestUniqueIdForEachSynthUser() {
	user1 := injector.NewSyntheticUser()
	user2 := injector.NewSyntheticUser()

	assert.NotEqual(s.T(), user1.Id(), user2.Id())
}

func TestSyntheticUserTestSuite(t *testing.T) {
	suite.Run(t, new(SyntheticUserTestSuite))
}
