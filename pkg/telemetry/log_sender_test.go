package telemetry_test

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
	"time"
)

type LogSenderTestSuite struct {
	suite.Suite
	bossMessenger   messaging.Messenger
	minionMessenger messaging.Messenger
}

func (l *LogSenderTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	l.bossMessenger, l.minionMessenger = messaging.NewChannelMessengerPair(100)
}

func (l *LogSenderTestSuite) TestNewLogSender() {
	dispatcher := telemetry.NewLogSender(l.minionMessenger, 10)

	if assert.IsType(l.T(), &telemetry.LogSender{}, dispatcher) {
		assert.Implements(l.T(), (*io.Writer)(nil), dispatcher)
	}
}

func (l *LogSenderTestSuite) TestWriteLogBelowBufferLimit() {
	dispatcher := telemetry.NewLogSender(l.minionMessenger, 10)
	logger := zerolog.New(dispatcher)
	logger.Info().Msg("my first log")
	logger.Info().Msg("my second log")

	// Check that no message was actually sent
	var noValue bool

	select {
	case <-l.bossMessenger.Receive():
		noValue = false

	default:
		noValue = true
	}

	assert.True(l.T(), noValue)
}

func (l *LogSenderTestSuite) TestWriteLogWithManualFlush() {
	dispatcher := telemetry.NewLogSender(l.minionMessenger, 10)
	logger := zerolog.New(dispatcher)
	logger.Info().Msg("my first log")
	logger.Info().Msg("my second log")

	dispatcher.Flush()
	var message messaging.Message

	select {
	case message = <-l.bossMessenger.Receive():

	default:
		assert.Fail(l.T(), "no message was sent")
	}

	if assert.IsType(l.T(), messaging.Message{}, message) {
		payload := message.Payload
		if assert.IsType(l.T(), &messaging.RemoteLogPayload{}, payload) {
			logLines := payload.(*messaging.RemoteLogPayload).Logs
			if assert.Len(l.T(), logLines, 2) {
				assert.Contains(l.T(), string(logLines[0]), "my first log")
				assert.Contains(l.T(), string(logLines[1]), "my second log")
			}
		}
	}
}

func (l *LogSenderTestSuite) TestWriteLogWithAutomaticFlush() {
	dispatcher := telemetry.NewLogSender(l.minionMessenger, 2)
	logger := zerolog.New(dispatcher)
	logger.Info().Msg("my first log")
	logger.Info().Msg("my second log")

	var message messaging.Message

	select {
	case message = <-l.bossMessenger.Receive():

	default:
		assert.Fail(l.T(), "no message was sent")
	}

	if assert.IsType(l.T(), messaging.Message{}, message) {
		payload := message.Payload
		if assert.IsType(l.T(), &messaging.RemoteLogPayload{}, payload) {
			logLines := payload.(*messaging.RemoteLogPayload).Logs
			if assert.Len(l.T(), logLines, 2) {
				assert.Contains(l.T(), string(logLines[0]), "my first log")
				assert.Contains(l.T(), string(logLines[1]), "my second log")
			}
		}
	}
}

func TestLogSenderTestSuite(t *testing.T) {
	suite.Run(t, new(LogSenderTestSuite))
}
