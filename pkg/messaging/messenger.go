package messaging

type Messenger interface {
	Receive() (Message, error)
	Send(Message) error
}
