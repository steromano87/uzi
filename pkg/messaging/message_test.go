package messaging_test

import (
	"github.com/gorilla/websocket"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTextMessage(t *testing.T) {
	message := messaging.NewMessage("TEST", map[string]any{"content": "test"})
	if assert.IsType(t, messaging.Message{}, message) {
		assert.Regexp(t, ".{8}-.{4}-.{4}-.{4}-.{12}", message.ID)
	}
}

func TestPackTextMessage(t *testing.T) {
	message := messaging.NewMessage("TEST", map[string]any{"content": "test"})
	messageType, packed, err := messaging.Pack(message)

	if assert.NoError(t, err) {
		assert.Equal(t, websocket.TextMessage, messageType)
		assert.IsType(t, []byte{}, packed)
	}
}

func TestUnpackTextMessage(t *testing.T) {
	message := messaging.NewMessage("TEST", map[string]any{"content": "test"})
	_, packed, _ := messaging.Pack(message)
	unpacked, err := messaging.Unpack(packed)

	if assert.NoError(t, err) {
		// TODO: avoid UnixNano time field comparison when bug https://github.com/stretchr/testify/issues/502 is fixed
		assert.Equal(t, message.ID, unpacked.ID)
		assert.Equal(t, message.Type, unpacked.Type)
		assert.Equal(t, message.Payload, unpacked.Payload)
		assert.Equal(t, message.Timestamp.UnixNano(), unpacked.Timestamp.UnixNano())
	}
}
