package telemetry_test

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/message"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
	"time"
)

type LogSenderTestSuite struct {
	suite.Suite
	bossMessenger   message.Bridge
	minionMessenger message.Bridge
}

func (l *LogSenderTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	l.bossMessenger, l.minionMessenger = message.NewChannelBridgePair(100)
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
	var msg *message.Envelope

	select {
	case msg = <-l.bossMessenger.Receive():

	default:
		assert.Fail(l.T(), "no msg was sent")
	}

	if assert.IsType(l.T(), &message.Envelope{}, msg) {
		payload := msg.GetLogs()

		if assert.NotNil(l.T(), payload) {
			if assert.IsType(l.T(), &message.Logs{}, payload) {
				logLines := payload.GetLog()
				if assert.Len(l.T(), logLines, 2) {
					assert.Contains(l.T(), string(logLines[0]), "my first log")
					assert.Contains(l.T(), string(logLines[1]), "my second log")
				}
			}
		}
	}
}

func (l *LogSenderTestSuite) TestWriteLogWithAutomaticFlush() {
	dispatcher := telemetry.NewLogSender(l.minionMessenger, 2)
	logger := zerolog.New(dispatcher)
	logger.Info().Msg("my first log")
	logger.Info().Msg("my second log")

	var msg *message.Envelope

	select {
	case msg = <-l.bossMessenger.Receive():

	default:
		assert.Fail(l.T(), "no msg was sent")
	}

	if assert.IsType(l.T(), &message.Envelope{}, msg) {
		payload := msg.GetLogs()

		if assert.NotNil(l.T(), payload) {
			if assert.IsType(l.T(), &message.Logs{}, payload) {
				logLines := payload.GetLog()
				if assert.Len(l.T(), logLines, 2) {
					assert.Contains(l.T(), string(logLines[0]), "my first log")
					assert.Contains(l.T(), string(logLines[1]), "my second log")
				}
			}
		}
	}
}

func TestLogSenderTestSuite(t *testing.T) {
	suite.Run(t, new(LogSenderTestSuite))
}
