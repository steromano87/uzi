package configuration

type Messaging struct {
	MessageCapacity int
	Logging         struct {
		BufferSize int
	}
	Samples struct {
		BufferSize int
	}
}
