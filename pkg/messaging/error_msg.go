package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
)

const ErrorMsgID = "ERROR"

type ErrorPayload struct {
	Message string `json:"message"`
}

func NewErrorMessage(err error) Message {
	return NewRawMessage(ErrorMsgID, &ErrorPayload{Message: err.Error()})
}

func NewErrorResponseMessage(originalMsgID string, err error) Message {
	return NewRawAnswerMessage(ErrorMsgID, originalMsgID, &ErrorPayload{Message: err.Error()})
}

func (e *ErrorPayload) UnmarshalJSON(bytes []byte) error {
	var intermediate map[string]any

	if err := json.Unmarshal(bytes, &intermediate); err != nil {
		return err
	}
	message, ok := intermediate[ErrorMsgID].(string)
	if !ok {
		return errors.New(fmt.Sprintf("error parsing message '%s' field as string: %x", ErrorMsgID, message))
	}

	e.Message = message
	return nil
}
