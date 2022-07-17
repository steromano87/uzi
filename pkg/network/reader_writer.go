package network

type Reader interface {
	Read() (int, []byte, error)
}

type Writer interface {
	Write(int, []byte) error
}

type ReadWriter interface {
	Reader
	Writer
}

type Messenger ReadWriter
