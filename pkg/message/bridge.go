package message

import (
	"context"
)

type Bridge interface {
	Start(ctx context.Context)
	Receive() <-chan *Envelope
	Send(msg *Envelope)
	SendPing()
	SendPong(pingMsgID string)
	Close()
}
