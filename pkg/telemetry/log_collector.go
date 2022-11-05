package telemetry

import (
	"sync"
)

type LogCollector struct {
	logs []*Log
	mu   sync.Mutex
}

func NewLogCollector() *LogCollector {
	collector := new(LogCollector)
	collector.logs = make([]*Log, 0)

	return collector
}

func (l *LogCollector) Write(p []byte) (n int, err error) {
	bytesToWrite := make([]byte, len(p))
	copy(bytesToWrite, p)
	l.AddLog(&Log{
		Entry: bytesToWrite,
	})

	return len(p), nil
}

func (l *LogCollector) AddLog(log *Log) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logs = append(l.logs, log)
}

func (l *LogCollector) GetLogs() []*Log {
	l.mu.Lock()
	defer l.mu.Unlock()

	outputLogs := make([]*Log, len(l.logs))
	copy(outputLogs, l.logs)
	l.logs = make([]*Log, 0)

	return outputLogs
}
