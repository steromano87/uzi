package telemetry

import (
	"github.com/steromano87/harkonnen/v1/pkg/message"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"sync"
)

type LogSender struct {
	messenger  messaging.Messenger
	bufferSize int

	queuedLogs [][]byte
	mu         sync.Mutex
}

func NewLogSender(messenger messaging.Messenger, bufferSize int) *LogSender {
	dispatcher := new(LogSender)
	dispatcher.messenger = messenger
	dispatcher.bufferSize = bufferSize
	dispatcher.queuedLogs = make([][]byte, 0)

	return dispatcher
}

func (l *LogSender) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	bytesToWrite := make([]byte, len(p))
	copy(bytesToWrite, p)
	l.queuedLogs = append(l.queuedLogs, bytesToWrite)
	l.mu.Unlock()

	if len(l.queuedLogs) >= l.bufferSize {
		l.Flush()
	}

	return len(p), nil
}

func (l *LogSender) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()

	payload := message.Envelope_Logs{Logs: &message.Logs{Log: l.queuedLogs}}
	msg := message.NewEnvelope(&payload)

	l.messenger.Send(msg)
	l.queuedLogs = make([][]byte, 0)
}
