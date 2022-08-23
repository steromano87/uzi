package messaging

import (
	"sync"
)

type ChannelMessenger struct {
	sendMu   sync.Mutex
	recvMu   sync.Mutex
	sendChan chan<- Message
	recvChan <-chan Message
}

func NewChannelMessenger(sendChan chan Message, recvChan chan Message) *ChannelMessenger {
	messenger := new(ChannelMessenger)
	messenger.sendChan = sendChan
	messenger.recvChan = recvChan

	return messenger
}

func NewChannelMessengerPair() (bossMessenger *ChannelMessenger, minionMessenger *ChannelMessenger) {
	bossToMinionChan := make(chan Message)
	minionToBossChan := make(chan Message)

	return NewChannelMessenger(bossToMinionChan, minionToBossChan), NewChannelMessenger(minionToBossChan, bossToMinionChan)
}

func (m *ChannelMessenger) Send(message Message) error {
	m.sendMu.Lock()
	defer m.sendMu.Unlock()
	m.sendChan <- message
	return nil
}

func (m *ChannelMessenger) Receive() (Message, error) {
	m.recvMu.Lock()
	defer m.recvMu.Unlock()

	return <-m.recvChan, nil
}

func (m *ChannelMessenger) SendPing() error {
	return m.Send(NewPingMessage())
}

func (m *ChannelMessenger) SendPong(pingMsgID string) error {
	return m.Send(NewPongMessage(pingMsgID))
}

func (m *ChannelMessenger) Close() {
	//TODO implement me
	panic("implement me")
}
