package messaging

import (
	"encoding/json"
)

const RemoteLogMsgId = "REMOTE_LOG"

type RemoteLogPayload struct {
	Logs []json.RawMessage `json:"logs"`
}

func NewRemoteLogMessage(logs []json.RawMessage) Message {
	return NewRawMessage(RemoteLogMsgId, &RemoteLogPayload{Logs: logs})
}

func (i *RemoteLogPayload) Type() string {
	return RemoteLogMsgId
}

func (i *RemoteLogPayload) UnmarshalJSON(bytes []byte) error {
	var intermediate map[string][]json.RawMessage

	if err := json.Unmarshal(bytes, &intermediate); err != nil {
		return err
	}
	i.Logs = intermediate[RemoteLogMsgId]

	return nil
}
