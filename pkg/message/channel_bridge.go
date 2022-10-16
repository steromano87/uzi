package message

import (
	"context"
)

type ChannelBridge struct {
	sendChan chan<- *Envelope
	recvChan <-chan *Envelope
}

func NewChannelBridge(sendChan chan *Envelope, recvChan chan *Envelope) *ChannelBridge {
	bridge := new(ChannelBridge)
	bridge.sendChan = sendChan
	bridge.recvChan = recvChan

	return bridge
}

func NewChannelBridgePair(channelBuffer int) (bossMessenger *ChannelBridge, minionMessenger *ChannelBridge) {
	bossToMinionChan := make(chan *Envelope, channelBuffer)
	minionToBossChan := make(chan *Envelope, channelBuffer)

	return NewChannelBridge(bossToMinionChan, minionToBossChan), NewChannelBridge(minionToBossChan, bossToMinionChan)
}

// Start is a no-op, since communication passes through native channels.
// No goroutines are used to pump data inside channels
func (m *ChannelBridge) Start(_ context.Context) {
}

func (m *ChannelBridge) Send(msg *Envelope) {
	m.sendChan <- msg
}

func (m *ChannelBridge) Receive() <-chan *Envelope {
	return m.recvChan
}

func (m *ChannelBridge) SendPing() {
	payload := Envelope_Ping{Ping: &Ping{}}
	msg := NewEnvelope(&payload)
	m.Send(msg)
}

func (m *ChannelBridge) SendPong(pingMsgID string) {
	payload := Envelope_Pong{Pong: &Pong{}}
	msg := NewResponseEnvelope(pingMsgID, &payload)
	m.Send(msg)
}

// Close is a no-op, because there is no goroutine to stop
func (m *ChannelBridge) Close() {
}
