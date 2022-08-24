package messaging

const GracefulShutdownMsgId = "GRACEFUL_SHUTDOWN"

func NewGracefulShutdownMessage() Message {
	return NewRawMessage(GracefulShutdownMsgId, nil)
}
