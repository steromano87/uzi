package telemetry_test

import (
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
	"time"
)

type LogCollectorTestSuite struct {
	suite.Suite
}

func (l *LogCollectorTestSuite) SetupTest() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
}

func (l *LogCollectorTestSuite) TestNewLogSender() {
	dispatcher := telemetry.NewLogCollector()

	if assert.IsType(l.T(), &telemetry.LogCollector{}, dispatcher) {
		assert.Implements(l.T(), (*io.Writer)(nil), dispatcher)
	}
}

func (l *LogCollectorTestSuite) TestWriteLogWithFlush() {
	dispatcher := telemetry.NewLogCollector()
	logger := zerolog.New(dispatcher)
	logger.Info().Msg("my first log")
	logger.Info().Msg("my second log")

	logs := dispatcher.GetLogs()

	if assert.IsType(l.T(), []*telemetry.Log{}, logs) {
		if assert.Len(l.T(), logs, 2) {
			assert.Contains(l.T(), string(logs[0].GetEntry()), "my first log")
			assert.Contains(l.T(), string(logs[1].GetEntry()), "my second log")
		}
	}
}

func TestLogCollectorTestSuite(t *testing.T) {
	suite.Run(t, new(LogCollectorTestSuite))
}
