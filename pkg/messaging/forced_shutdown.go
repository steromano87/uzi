package messaging

const ForcedShutdownMsgId = "FORCED_SHUTDOWN"

func NewForcedShutdownMessage() Message {
	return NewRawMessage(ForcedShutdownMsgId, nil)
}
