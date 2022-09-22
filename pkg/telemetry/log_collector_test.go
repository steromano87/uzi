package telemetry_test

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/db"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type LogCollectorTestSuite struct {
	suite.Suite
	ctx             context.Context
	bossMessenger   messaging.Messenger
	minionMessenger messaging.Messenger
	dbAdapter       *db.Adapter
}

func (s *LogCollectorTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro

	s.ctx = context.TODO()
	s.bossMessenger, s.minionMessenger = messaging.NewChannelMessengerPair(100)
	s.dbAdapter = db.NewAdapter(db.SQLite, "file::memory:?cache=shared")
	_ = s.dbAdapter.Connect(s.ctx)
}

func (s *LogCollectorTestSuite) TestNewLogCollector() {
	collector := telemetry.NewLogCollector("myInjector", s.dbAdapter)

	assert.IsType(s.T(), telemetry.LogCollector{}, collector)
}

func (s *LogCollectorTestSuite) TestCollectMultipleLogs() {
	collector := telemetry.NewLogCollector("myInjector", s.dbAdapter)

	dispatcher := telemetry.NewLogSender(s.minionMessenger, 10)
	logger := zerolog.New(dispatcher).With().Timestamp().Logger()
	logger.Info().Msg("my first log")
	logger.Info().Msg("my second log")

	dispatcher.Flush()
	var message messaging.Message

	select {
	case message = <-s.bossMessenger.Receive():

	default:
		assert.Fail(s.T(), "no message was sent")
	}

	payload := message.Payload.(*messaging.RemoteLogPayload)

	err := collector.Collect(payload)
	if assert.NoError(s.T(), err) {
		var logs []db.Log
		s.dbAdapter.Find(&logs)

		if assert.Len(s.T(), logs, 2) {
			assert.Equal(s.T(), logs[0].Level, "info")
			assert.Equal(s.T(), logs[0].Message, "my first log")

			assert.Equal(s.T(), logs[1].Level, "info")
			assert.Equal(s.T(), logs[1].Message, "my second log")
		}

	}
}

func TestLogCollectorTestSuite(t *testing.T) {
	suite.Run(t, new(LogCollectorTestSuite))
}
