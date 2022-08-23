package messaging_test

import (
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRunnerQuotaUpdateMessage(t *testing.T) {
	message := messaging.NewRunnerQuotaUpdateMessage(5)

	if assert.IsType(t, messaging.Message{}, message) {
		assert.Equal(t, messaging.RunnersQuotaUpdateMsgId, message.Type)
		payload, ok := message.Payload.(*messaging.RunnersQuotaUpdatePayload)

		if assert.True(t, ok) {
			assert.Equal(t, 5, payload.Quota)
		}
	}
}
