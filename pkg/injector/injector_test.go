package injector_test

import (
	"context"
	"github.com/Flaque/filet"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/protobuf/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"path/filepath"
	"testing"
	"time"
)

type InjectorTestSuite struct {
	suite.Suite
	bossMessenger   messaging.Messenger
	minionMessenger messaging.Messenger
	ctx             injector.Context
	cancelFunc      context.CancelFunc
}

func (s *InjectorTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	s.bossMessenger, s.minionMessenger = messaging.NewChannelMessengerPair(100)
	s.ctx, s.cancelFunc = injector.NewContext(context.TODO(), &logger, s.minionMessenger)
}

func (s *InjectorTestSuite) TearDownTest() {
	s.cancelFunc()
}

func (s *InjectorTestSuite) TestNewInjector() {
	inj, err := injector.New(s.ctx)
	defer inj.Stop()

	if assert.NoError(s.T(), err) {
		workingFolder := inj.WorkingFolder()
		assert.DirExists(s.T(), workingFolder)
	}
}

func (s *InjectorTestSuite) TestStartNewInjector() {
	inj, err := injector.New(s.ctx)
	if assert.NoError(s.T(), err) {
		assert.Equal(s.T(), injector.Stopped, inj.Status())
		inj.Start()
		assert.Equal(s.T(), injector.Ready, inj.Status())
	}
}

func (s *InjectorTestSuite) TestPingMessageHandling() {
	inj, _ := injector.New(s.ctx)
	inj.Start()
	s.bossMessenger.SendPing()

	responseMessage, ok := <-s.bossMessenger.Receive()
	if assert.True(s.T(), ok) {
		assert.IsType(s.T(), &message.Envelope_Pong{}, responseMessage.GetPayload())
		assert.NotNil(s.T(), responseMessage.GetAnswersTo())
	}
}

func (s *InjectorTestSuite) TestWorkingFolderInitMessageHandling() {
	inj, _ := injector.New(s.ctx)
	inj.Start()

	tempWorkingDir := filet.TmpDir(s.T(), "")
	tempFile := filet.TmpFile(s.T(), tempWorkingDir, "sample")

	defer filet.CleanUp(s.T())

	workDirMessage, err := message.NewWorkingFolderInitEnvelope(tempWorkingDir)
	require.NoError(s.T(), err)

	s.bossMessenger.Send(workDirMessage)

	responseMessage, ok := <-s.bossMessenger.Receive()
	if assert.True(s.T(), ok) {
		if assert.IsType(s.T(), &message.Envelope_Acknowledge{}, responseMessage.GetPayload()) {
			assert.Equal(s.T(), workDirMessage.GetId(), responseMessage.GetAnswersTo())
			assert.True(s.T(), responseMessage.GetAcknowledge().GetOk())
		}

		if assert.DirExists(s.T(), inj.WorkingFolder()) {
			assert.FileExists(s.T(), filepath.Join(inj.WorkingFolder(), tempFile.Name()))
		}
	}
}

func TestInjectorTestSuite(t *testing.T) {
	suite.Run(t, new(InjectorTestSuite))
}
