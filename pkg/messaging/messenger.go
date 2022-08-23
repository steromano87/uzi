package messaging

type Messenger interface {
	Receive() (Message, error)
	Send(Message) error
	SendPing() error
	SendPong(pingMsgID string) error
	Close()
}
