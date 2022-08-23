package messaging

const (
	PingMsgId = "PING"
	PongMsgId = "PONG"
)

func NewPingMessage() Message {
	return NewRawMessage(PingMsgId, nil)
}

func NewPongMessage(pingMsgID string) Message {
	return NewRawAnswerMessage(PongMsgId, pingMsgID, nil)
}
