package configuration

type MessagingConfiguration struct {
	MessageCapacity int
	Logging         struct {
		BufferSize int
	}
	Samples struct {
		BufferSize int
	}
}
