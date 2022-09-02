package injector

import (
	"encoding/json"
	"github.com/steromano87/harkonnen/v1/pkg/messaging"
	"sync"
)

type LogSender struct {
	messenger  messaging.Messenger
	bufferSize int

	queuedLogs []json.RawMessage
	mu         sync.Mutex
}

func NewLogSender(messenger messaging.Messenger, bufferSize int) *LogSender {
	dispatcher := new(LogSender)
	dispatcher.messenger = messenger
	dispatcher.bufferSize = bufferSize
	dispatcher.queuedLogs = make([]json.RawMessage, 0)

	return dispatcher
}

func (l *LogSender) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.queuedLogs = append(l.queuedLogs, p)

	if len(l.queuedLogs) >= l.bufferSize {
		l.Flush()
	}

	return len(p), nil
}

func (l *LogSender) Flush() {
	logMessage := messaging.NewRemoteLogMessage(l.queuedLogs)
	l.messenger.Send(logMessage)
	l.queuedLogs = make([]json.RawMessage, 0)
}
