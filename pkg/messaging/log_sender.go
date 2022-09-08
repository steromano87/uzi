package messaging

import (
	"encoding/json"
	"sync"
)

type LogSender struct {
	messenger  Messenger
	bufferSize int

	queuedLogs []json.RawMessage
	mu         sync.Mutex
}

func NewLogSender(messenger Messenger, bufferSize int) *LogSender {
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
	logMessage := NewRemoteLogMessage(l.queuedLogs)
	l.messenger.Send(logMessage)
	l.queuedLogs = make([]json.RawMessage, 0)
}
