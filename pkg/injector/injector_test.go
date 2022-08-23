package injector_test

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/injector"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type InjectorTestSuite struct {
	suite.Suite
	bossMessenger   messaging.Messenger
	minionMessenger messaging.Messenger
	ctx             messaging.Context
	cancelFunc      context.CancelFunc
}

func (s *InjectorTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = "2006-01-02T15:04:05.000"
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	s.bossMessenger, s.minionMessenger = messaging.NewChannelMessengerPair()
	s.ctx, s.cancelFunc = messaging.NewContext(context.TODO(), &logger, s.minionMessenger)
}

func (s *InjectorTestSuite) TearDownTest() {
	s.cancelFunc()
}

func (s *InjectorTestSuite) TestStartNewInjector() {
	inj := injector.New(s.ctx)
	if assert.Equal(s.T(), injector.Stopped, inj.Status()) {
		inj.Run()
		assert.Equal(s.T(), injector.Ready, inj.Status())
	}
}

func (s *InjectorTestSuite) TestMessageHandling() {
	inj := injector.New(s.ctx)
	inj.Run()
	err := s.bossMessenger.SendPing()

	if assert.NoError(s.T(), err) {
		responseMessage, err := s.bossMessenger.Receive()
		if assert.NoError(s.T(), err) {
			assert.Equal(s.T(), messaging.PongMsgId, responseMessage.Type)
			assert.NotEmpty(s.T(), responseMessage.AnswersTo)
		}
	}
}

func TestInjectorTestSuite(t *testing.T) {
	suite.Run(t, new(InjectorTestSuite))
}
