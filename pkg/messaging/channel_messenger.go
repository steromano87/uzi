package messaging

import (
	"context"
	"github.com/steromano87/harkonnen/v1/pkg/message"
)

type ChannelMessenger struct {
	sendChan chan<- *message.Envelope
	recvChan <-chan *message.Envelope
}

func NewChannelMessenger(sendChan chan *message.Envelope, recvChan chan *message.Envelope) *ChannelMessenger {
	messenger := new(ChannelMessenger)
	messenger.sendChan = sendChan
	messenger.recvChan = recvChan

	return messenger
}

func NewChannelMessengerPair(channelBuffer int) (bossMessenger *ChannelMessenger, minionMessenger *ChannelMessenger) {
	bossToMinionChan := make(chan *message.Envelope, channelBuffer)
	minionToBossChan := make(chan *message.Envelope, channelBuffer)

	return NewChannelMessenger(bossToMinionChan, minionToBossChan), NewChannelMessenger(minionToBossChan, bossToMinionChan)
}

// Start is a no-op, since communication passes through native channels.
// No goroutines are used to pump data inside channels
func (m *ChannelMessenger) Start(_ context.Context) {
}

func (m *ChannelMessenger) Send(msg *message.Envelope) {
	m.sendChan <- msg
}

func (m *ChannelMessenger) Receive() <-chan *message.Envelope {
	return m.recvChan
}

func (m *ChannelMessenger) SendPing() {
	payload := message.Envelope_Ping{Ping: &message.Ping{}}
	msg := message.NewEnvelope(&payload)
	m.Send(msg)
}

func (m *ChannelMessenger) SendPong(pingMsgID string) {
	payload := message.Envelope_Pong{Pong: &message.Pong{}}
	msg := message.NewResponseEnvelope(pingMsgID, &payload)
	m.Send(msg)
}

// Close is a no-op, because there is no goroutine to stop
func (m *ChannelMessenger) Close() {
}
