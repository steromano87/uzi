package messaging

import (
	"context"
)

type ChannelMessenger struct {
	sendChan chan<- Message
	recvChan <-chan Message
}

func NewChannelMessenger(sendChan chan Message, recvChan chan Message) *ChannelMessenger {
	messenger := new(ChannelMessenger)
	messenger.sendChan = sendChan
	messenger.recvChan = recvChan

	return messenger
}

func NewChannelMessengerPair(channelBuffer int) (bossMessenger *ChannelMessenger, minionMessenger *ChannelMessenger) {
	bossToMinionChan := make(chan Message, channelBuffer)
	minionToBossChan := make(chan Message, channelBuffer)

	return NewChannelMessenger(bossToMinionChan, minionToBossChan), NewChannelMessenger(minionToBossChan, bossToMinionChan)
}

// Start is a no-op, since communication passes through native channels.
// No goroutines are used to pump data inside channels
func (m *ChannelMessenger) Start(_ context.Context) {
}

func (m *ChannelMessenger) Send(message Message) {
	m.sendChan <- message
}

func (m *ChannelMessenger) Receive() <-chan Message {
	return m.recvChan
}

func (m *ChannelMessenger) SendPing() {
	m.Send(NewPingMessage())
}

func (m *ChannelMessenger) SendPong(pingMsgID string) {
	m.Send(NewPongMessage(pingMsgID))
}

// Close is a no-op, because there is no goroutine to stop
func (m *ChannelMessenger) Close() {
}
