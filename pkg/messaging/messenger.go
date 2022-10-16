package messaging

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/message"
)

type Messenger interface {
	Start(ctx context.Context)
	Receive() <-chan *message.Envelope
	Send(msg *message.Envelope)
	SendPing()
	SendPong(pingMsgID string)
	Close()
}
