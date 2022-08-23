package messaging_test

import (
	"bytes"
	"github.com/rs/zerolog"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRemoteLogMessage(t *testing.T) {
	var buffer bytes.Buffer
	logger := zerolog.New(&buffer)
	logger.Info().Str("randomKey", "randomValue").Msg("First log")
	logger.Warn().Int("randomInt", 1).Msg("Second log")

	message := messaging.NewRemoteLogMessage(buffer.Bytes())

	if assert.IsType(t, messaging.Message{}, message) {
		assert.Equal(t, messaging.RemoteLogMsgId, message.Type)
		payload, ok := message.Payload.(*messaging.RemoteLogPayload)

		if assert.True(t, ok) {
			assert.EqualValues(t, buffer.Bytes(), payload.Logs)
		}
	}
}
