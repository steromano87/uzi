package configuration

type MessagingConfiguration struct {
	MessageCapacity int `default:"2048"`
	Logging         struct {
		BufferSize int `default:"50"`
	}
	Samples struct {
		BufferSize int `default:"10"`
	}
}
