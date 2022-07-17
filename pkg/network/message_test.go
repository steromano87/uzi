package network_test

import (
	"github.com/gorilla/websocket"
	"github.com/steromano87/harkonnen/v1/pkg/network"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestNewTextMessage(t *testing.T) {
	message := network.NewMessage("TEST", map[string]string{"content": "test"})
	if assert.IsType(t, &network.Message{}, message) {
		assert.Regexp(t, ".{8}-.{4}-.{4}-.{4}-.{12}", message.ID)
		assert.Equal(t, websocket.TextMessage, message.WebSocketType())
	}
}

func TestNewBinaryMessage(t *testing.T) {
	bytes := make([]byte, 20)
	rand.Read(bytes)
	message := network.NewMessage("TEST", bytes)
	if assert.IsType(t, &network.Message{}, message) {
		assert.Regexp(t, ".{8}-.{4}-.{4}-.{4}-.{12}", message.ID)
		assert.Equal(t, websocket.BinaryMessage, message.WebSocketType())
	}
}

func TestPackTextMessage(t *testing.T) {
	message := network.NewMessage("TEST", map[string]string{"content": "test"})
	messageType, packed, err := network.PackMessage(message)

	if assert.NoError(t, err) {
		assert.Equal(t, websocket.TextMessage, messageType)
		assert.IsType(t, []byte{}, packed)
	}
}

func TestPackBinaryMessage(t *testing.T) {
	bytes := make([]byte, 20)
	rand.Read(bytes)
	message := network.NewMessage("TEST", bytes)
	messageType, packed, err := network.PackMessage(message)

	if assert.NoError(t, err) {
		assert.Equal(t, websocket.BinaryMessage, messageType)
		assert.IsType(t, []byte{}, packed)
	}
}

func TestUnpackTextMessage(t *testing.T) {
	message := network.NewMessage("TEST", map[string]any{"content": "test"})
	_, packed, _ := network.PackMessage(message)
	unpacked, err := network.UnpackMessage(packed)

	if assert.NoError(t, err) {
		// TODO: avoid UnixNano time field comparison when bug https://github.com/stretchr/testify/issues/502 is fixed
		assert.Equal(t, message.ID, unpacked.ID)
		assert.Equal(t, message.Type, unpacked.Type)
		assert.Equal(t, message.Payload, unpacked.Payload)
		assert.Equal(t, message.Timestamp.UnixNano(), unpacked.Timestamp.UnixNano())
	}
}
