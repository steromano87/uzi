package network

import (
	"encoding/json"
)

func PackMessage(message *Message) (int, []byte, error) {
	marshalledMessage, err := json.Marshal(message)
	if err != nil {
		return 0, nil, err
	}

	return message.WebSocketType(), marshalledMessage, nil
}

func UnpackMessage(packedMessage []byte) (*Message, error) {
	var unmarshalledMessage Message
	err := json.Unmarshal(packedMessage, &unmarshalledMessage)
	if err != nil {
		return nil, err
	}

	return &unmarshalledMessage, nil
}
