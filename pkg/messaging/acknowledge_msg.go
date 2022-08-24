package messaging

const AcknowledgeMsgID = "ACK"

func NewAcknowledgeMessage(originalMsgID string) Message {
	return NewRawAnswerMessage(AcknowledgeMsgID, originalMsgID, nil)
}
