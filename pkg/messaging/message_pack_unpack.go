package messaging

import (
	"encoding/json"
	"github.com/gorilla/websocket"
)

func Pack(message Message) (int, []byte, error) {
	marshalledMessage, err := json.Marshal(message)
	if err != nil {
		return 0, nil, err
	}

	return websocket.TextMessage, marshalledMessage, nil
}

func Unpack(packedMessage []byte) (Message, error) {
	var unmarshalledMessage Message
	err := json.Unmarshal(packedMessage, &unmarshalledMessage)
	if err != nil {
		return Message{}, err
	}

	return unmarshalledMessage, nil
}
