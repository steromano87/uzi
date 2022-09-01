package messaging_test

import (
	"bytes"
	"encoding/json"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRemoteLogMessage(t *testing.T) {
	var buffer bytes.Buffer
	rawLogs := make([]json.RawMessage, 0)
	logger := zerolog.New(&buffer)
	logger.Info().Str("randomKey", "randomValue").Msg("First log")
	rawLogs = append(rawLogs, buffer.Bytes())
	buffer.Reset()

	logger.Warn().Int("randomInt", 1).Msg("Second log")
	rawLogs = append(rawLogs, buffer.Bytes())
	buffer.Reset()

	message := messaging.NewRemoteLogMessage(rawLogs)

	if assert.IsType(t, messaging.Message{}, message) {
		assert.Equal(t, messaging.RemoteLogMsgId, message.Type)
		payload, ok := message.Payload.(*messaging.RemoteLogPayload)

		if assert.True(t, ok) {
			assert.EqualValues(t, rawLogs, payload.Logs)
		}
	}
}
