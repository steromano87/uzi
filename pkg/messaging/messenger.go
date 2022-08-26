package messaging

import "context"

type Messenger interface {
	Start(ctx context.Context)
	Receive() <-chan Message
	Send(message Message)
	SendPing()
	SendPong(pingMsgID string)
	Close()
}
